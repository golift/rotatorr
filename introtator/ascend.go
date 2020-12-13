package introtator

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// rotateAscending handles the rotation of integer log files. Integers just means
// a bare number is appended to the name. In the default, ascending mode, the previous
// log file is always .1.log. In descending mode, the previous log is always the highest
// number. In the default ascending mode, every log file has to be renamed on every rotation.
func (l *Layout) rotateAscending(logFiles *backupFiles, fileName string) (string, error) {
	var (
		new     = 1
		dir     = l.getArchiveDir(fileName)
		prefix  = l.getPrefix(fileName)
		newPath = filepath.Join(dir, prefix+LogExt1)
	)

	if len(logFiles.Files) != 0 {
		// ascending and we have files. They all need to be renamed.
		for i, f := range logFiles.Files {
			ext := LogExt
			if strings.HasSuffix(logFiles.Files[i], GZext) {
				ext += GZext
			}

			if i != len(logFiles.Files)-1 && logFiles.value[i+1] != logFiles.value[i]-1 {
				continue // There's a gap in the list, so skip renaming one.
			}

			logFiles.value[i]++
			logFiles.Files[i] = filepath.Join(dir, prefix+strconv.Itoa(logFiles.value[i])+ext)

			// fmt.Printf("\nrenaming [%d] %s -> %s\n", i, f, logFiles.Files[i])
			if err := l.Rename(f, logFiles.Files[i]); err != nil {
				return "", fmt.Errorf("error rotating backup file: %w", err)
			}
		}
	}

	if err := l.Rename(fileName, newPath); err != nil {
		return "", fmt.Errorf("error rotating file: %w", err)
	}

	logFiles.Files = append(logFiles.Files, newPath)
	logFiles.value = append(logFiles.value, new)

	return newPath, nil
}

// deleteOldLogsAsc deletes old files based on max file count.
func (l *Layout) deleteOldLogsAsc(logFiles *backupFiles) error {
	if l.FileCount < 1 {
		return nil
	}

	count := len(logFiles.Files)

	for _, f := range logFiles.Files {
		if count <= l.FileCount {
			break
		}

		if err := l.Remove(f); err != nil {
			return fmt.Errorf("error removing file: %w", err)
		}

		count--
	}

	// fmt.Println("deleted:", files)
	return nil
}
