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
		dir     = l.getArchiveDir(fileName)
		prefix  = l.getPrefix(fileName)
		newPath = filepath.Join(dir, prefix+LogExt1)
	)

	if len(logFiles.Files) != 0 {
		// ascending and we have files. They all need to be renamed.
		for idx, filePath := range logFiles.Files {
			ext := LogExt
			if strings.HasSuffix(logFiles.Files[idx], GZext) {
				ext += GZext
			}

			if idx != len(logFiles.Files)-1 && logFiles.value[idx+1] != logFiles.value[idx]-1 {
				continue // There's a gap in the list, so skip renaming one.
			}

			logFiles.value[idx]++
			logFiles.Files[idx] = filepath.Join(dir, prefix+strconv.Itoa(logFiles.value[idx])+ext)

			// fmt.Printf("\nrenaming [%d] %s -> %s\n", i, f, logFiles.Files[i])
			err := l.Rename(filePath, logFiles.Files[idx])
			if err != nil {
				return "", fmt.Errorf("error rotating backup file: %w", err)
			}
		}
	}

	err := l.Rename(fileName, newPath)
	if err != nil {
		return "", fmt.Errorf("error rotating file: %w", err)
	}

	logFiles.Files = append(logFiles.Files, newPath)
	logFiles.value = append(logFiles.value, 1)

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

		err := l.Remove(f)
		if err != nil {
			return fmt.Errorf("error removing file: %w", err)
		}

		count--
	}

	// fmt.Println("deleted:", files)
	return nil
}
