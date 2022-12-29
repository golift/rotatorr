// Package introtator provides an interface for Rotatorr that renames
// backup log files with an incrementing integer in the name.
// By default rotated log files are named: service.1.log.
//
// Control the order of integers with Layout.FileOrder.
//
// In Ascending mode, the current file is always rotated to `.1`, and
// all the files in the way are rotated first. In Descending mode the
// existing files are pruned, then files are rotated down and lastly
// the current file is rotated to the next highest integer.
package introtator

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golift.io/rotatorr"
	"golift.io/rotatorr/filer"
)

// Order defines which direction the files are written in.
type Order uint8

// Ascending order writes files 1-10 in sequence.
// Descending order writes them 10-1 in sequence.
const (
	Ascending Order = iota
	Descending
)

// Layout defines how integer-stamped backup logs have their file names decided.
// This also sets how many files are kept; default is unlimited. Recommend setting
// FileCount when FileOrder is set to Ascending (default), otherwise the app may
// spend a lot of time renaming files. If you enable compression with a PostRotate
// hook, make sure compression finishes before the files are rotated.
type Layout struct {
	ArchiveDir string // Location where rotated backup logs are moved to.
	FileCount  int    // Maximum number of rotated log files.
	FileOrder  Order  // Control the order of the integer-named backup log files.
	PostRotate func(fileName, newFile string)
	filer.Filer
}

// Some constant this package uses.
const (
	LogExt  = ".log"  // suffixed to an integer.
	LogExt1 = "1.log" // suffixed to the prefix.
	GZext   = ".gz"   // trimmed off found files.
	Joiner  = "."     // joins prefix with integer.
)

// Rotate forces the log to rotate immediately. Returns the new name of the rotated log.
func (l *Layout) Rotate(fileName string) (string, error) {
	switch logFiles := l.getAllLogFiles(fileName); l.FileOrder {
	case Descending:
		sort.Sort(logFiles)

		remainingfiles, err := l.deleteOldLogsDesc(logFiles)
		if err != nil {
			return "", err
		}

		return l.rotateDescending(remainingfiles, fileName)
	case Ascending:
		fallthrough
	default:
		sort.Sort(sort.Reverse(logFiles))

		newFile, err := l.rotateAscending(logFiles, fileName)
		if err != nil {
			return newFile, err
		}

		return newFile, l.deleteOldLogsAsc(logFiles)
	}
}

// Dirs checks our config and returns the folder for rotatorr library to create them.
func (l *Layout) Dirs(fileName string) ([]string, error) {
	if l.Filer == nil {
		l.Filer = filer.Default()
	}

	if l.FileOrder > Descending {
		l.FileOrder = Ascending
	}

	switch fpath := filepath.Dir(fileName); {
	case l.ArchiveDir == "" || fpath == l.ArchiveDir:
		return []string{fpath}, nil
	default:
		return []string{fpath, l.ArchiveDir}, nil
	}
}

// Post satisfies the Rotatorr interface.
func (l *Layout) Post(fileName, newFile string) {
	if l.PostRotate != nil {
		l.PostRotate(fileName, newFile)
	}
}

// GetPrefix returns a file's prefix. Removes the path and extension.
// This is used internally, but exposed for convenience when writing your own logic.
func (l *Layout) getPrefix(fileName string) string {
	return strings.TrimSuffix(filepath.Base(fileName), LogExt) + Joiner
}

// GetArchiveDir returns the archive directory if one is set,
// otherwise the directory the log file is in.
func (l *Layout) getArchiveDir(fileName string) string {
	if l.ArchiveDir != "" {
		return l.ArchiveDir
	}

	return filepath.Dir(fileName)
}

// GetAllLogFiles finds all the backup log files that match our pattern.
func (l *Layout) getAllLogFiles(fileName string) *backupFiles {
	var (
		dir    = l.getArchiveDir(fileName)
		list   = &backupFiles{Files: []string{}, value: []int{}}
		prefix = l.getPrefix(fileName)
	)

	files, err := l.ReadDir(dir)
	if err != nil || len(files) == 0 {
		return list
	}

	for _, file := range files {
		name := file.Name()
		if !strings.HasPrefix(name, prefix) {
			continue // not our file.
		}

		part := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(name, prefix), GZext), LogExt)

		i, err := strconv.Atoi(part)
		if err == nil {
			list.Files = append(list.Files, filepath.Join(dir, name))
			list.value = append(list.value, i)
		}
	}

	// fmt.Println("file list:", list)
	return list
}

// Our interface must satify a rotatorr.Rotatorr.
var _ rotatorr.Rotatorr = (*Layout)(nil)
