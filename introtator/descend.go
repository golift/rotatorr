package introtator

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// rotate handles the rotation of integer log files. Integers just means
// a bare number is appended to the name. In the default, ascending mode, the previous
// log file is always .1.log. In descending mode, the previous log is always the highest
// number. In the default ascending mode, every log file has to be renamed on every rotation.
func (l *Layout) rotateDescending(logFiles *backupFiles, fileName string) (string, error) {
	var (
		dir    = l.getArchiveDir(fileName)
		prefix = l.getPrefix(fileName)
	)

	for idx, filePath := range logFiles.Files {
		ext := LogExt
		if strings.HasSuffix(logFiles.Files[idx], GZext) {
			ext += GZext
		}

		logFiles.value[idx] = idx + 1
		logFiles.Files[idx] = filepath.Join(dir, prefix+strconv.Itoa(logFiles.value[idx])+ext)

		if logFiles.Files[idx] == filePath {
			continue // this shouldn't happen.
		}

		// fmt.Printf("\nrenaming [%d] %s -> %s\n", i, f, logFiles.Files[i])
		if err := l.Rename(filePath, logFiles.Files[idx]); err != nil {
			return "", fmt.Errorf("error rotating file: %w", err)
		}
	}

	newPath := filepath.Join(dir, prefix+strconv.Itoa(len(logFiles.value)+1)+LogExt)

	if err := l.Rename(fileName, newPath); err != nil {
		return "", fmt.Errorf("error rotating file: %w", err)
	}

	logFiles.Files = append(logFiles.Files, newPath)
	logFiles.value = append(logFiles.value, len(logFiles.value)+1)

	return newPath, nil
}

// deleteOldLogsDesc deletes old files based on max file count.
func (l *Layout) deleteOldLogsDesc(logFiles *backupFiles) (*backupFiles, error) {
	files := &backupFiles{Files: []string{}, value: []int{}}
	count := len(logFiles.Files)

	if l.FileCount < 1 {
		return files, nil
	}

	for idx, filePath := range logFiles.Files {
		if count < l.FileCount {
			files.Files = append(files.Files, filePath)
			files.value = append(files.value, logFiles.value[idx])

			continue
		}

		// fmt.Println("deleted", filePath, count)
		if err := l.Remove(filePath); err != nil {
			return files, fmt.Errorf("error removing file: %w", err)
		}

		count--
	}

	// fmt.Println("kept:", files)
	return files, nil
}
