// Package timerotator provides an interface for Rotatorr that renames backup
// log files with a time stamp in the name. This package provides the ability
// to limit backup log files by count (number of logs) and by age (of files).
// By default rotated log files are named: service-2006-01-02T15-04-05.000.log.
// Control the time format with the Layout.Format parameter. The defaults in this
// package work very similarly to: https://github.com/natefinch/lumberjack
package timerotator

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golift.io/rotatorr"
	"golift.io/rotatorr/filer"
)

// Layout defines how time-stamped backup logs have their file names decided.
type Layout struct {
	filer.Filer

	ArchiveDir string        // Location where rotated backup logs are moved to.
	FileCount  int           // Maximum number of rotated log files.
	FileAge    time.Duration // Maximum age of rotated files.
	UseUTC     bool          // Sets the time zone to UTC when writing Time Formats (backup files).
	Format     string        // Format for Go Time. Used as the name.
	Joiner     string        // The string betwene the file name prefix and time stamp. Default: -
	// Mockable interfaces. Can be used for custom processing. Setting these is very optional.
	PostRotate func(fileName, newFile string)
}

// Some Formats you may use in your app.
const (
	FormatDefault = "2006-01-02T15-04-05.000" // Default: Used if Format = ""
	FormatNoSecnd = "2006-01-02T15-04-05"     // Example: Same thing, sans msec.
	FormatDumbUSA = "02-01-2006_15:04:05"     // Example: Silly Americans.
)

// Some constant this package uses; not really needed externally.
const (
	LogExt        = ".log"
	DefaultJoiner = "-"
	GZext         = ".gz"
)

// Post satisfies the Rotatorr interface.
func (l *Layout) Post(fileName, newFile string) {
	if l.PostRotate != nil {
		l.PostRotate(fileName, newFile)
	}
}

// Rotate forces the log to rotate immediately. Returns the size of the rotated log.
func (l *Layout) Rotate(fileName string) (string, error) {
	now := time.Now()
	if l.UseUTC {
		now = now.UTC()
	}

	var (
		dir     = l.getArchiveDir(fileName)
		newFile = filepath.Join(dir, l.getPrefix(fileName)+now.Format(l.Format)+LogExt)
	)

	err := l.Rename(fileName, newFile)
	if err != nil {
		return "", fmt.Errorf("error renaming log: %w", err)
	}

	return newFile, l.deleteOldLogs(l.getAllLogFiles(fileName))
}

// Dirs validates input data and returns the list of directories being used.
func (l *Layout) Dirs(fileName string) ([]string, error) {
	if l.Format == "" {
		l.Format = FormatDefault
	}

	if l.Joiner == "" {
		l.Joiner = DefaultJoiner
	}

	if l.Filer == nil {
		l.Filer = filer.Default()
	}

	switch fpath := filepath.Dir(fileName); {
	case l.ArchiveDir == "" || fpath == l.ArchiveDir:
		return []string{fpath}, nil
	default:
		return []string{fpath, l.ArchiveDir}, nil
	}
}

func (l *Layout) getArchiveDir(fileName string) string {
	if l.ArchiveDir != "" {
		return l.ArchiveDir
	}

	return filepath.Dir(fileName)
}

// deleteOldLogs deletes any files that are older than FileAge (if Format=Time).
// Then it deletes extra logs if we're over our NumFiles count.
func (l *Layout) deleteOldLogs(logFiles *backupFiles) error {
	gone := make(map[string]struct{})

	if l.FileAge > 0 {
		// Parse the time stamp out of each file name.
		// If the time is older than FileAge, delete the file.
		for idx, when := range logFiles.value {
			if time.Since(when) < l.FileAge {
				continue
			}

			err := l.Remove(logFiles.Files[idx])
			if err != nil {
				return fmt.Errorf("error removing file: %w", err)
			}

			gone[logFiles.Files[idx]] = struct{}{}
		}
	}

	count := len(logFiles.Files) - len(gone)

	if l.FileCount > 0 {
		for _, fileName := range logFiles.Files {
			if count <= l.FileCount {
				return nil
			}

			if _, ok := gone[fileName]; ok {
				continue // already deleted this one.
			}

			err := l.Remove(fileName)
			if err != nil {
				return fmt.Errorf("error removing file: %w", err)
			}

			count--
		}
	}

	return nil
}

// getPrefix returns the expected - or created - prefix on our log files.
func (l *Layout) getPrefix(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), LogExt) + l.Joiner
}

// getAllLogFiles finds all the backup log files that match our Time Format.
func (l *Layout) getAllLogFiles(fileName string) *backupFiles {
	var (
		list   = &backupFiles{Files: []string{}, value: []time.Time{}}
		dir    = l.getArchiveDir(fileName)
		prefix = l.getPrefix(fileName)
	)

	fileList, err := l.ReadDir(dir)
	if err != nil || len(fileList) == 0 {
		return list
	}

	for idx := range fileList {
		name := fileList[idx].Name()
		if !strings.HasPrefix(name, prefix) {
			continue // not our file.
		}

		part := strings.TrimSuffix(strings.TrimPrefix(name, prefix), GZext)

		t, err := time.Parse(l.Format, strings.TrimSuffix(part, LogExt))
		if err == nil { // if err != nil, then not our file.
			list.Files = append(list.Files, filepath.Join(dir, name))
			list.value = append(list.value, t)
		}
	}

	sort.Sort(list)

	// fmt.Println(time.Now(), "PREFIX:", prefix, "FILES:", list)
	return list
}

// Our interface must satify a rotatorr.Rotatorr.
var _ rotatorr.Rotatorr = (*Layout)(nil)
