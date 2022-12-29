package filer

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Stat returns a *FileInfo struct w/ attached os.FileInfo interface.
func Stat(filename string) (*FileInfo, error) {
	fileStat, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("stat err: %w", err)
	}

	var unixTime int64

	sysCtime, _ := fileStat.Sys().(*syscall.Win32FileAttributeData)
	if sysCtime != nil {
		unixTime = sysCtime.CreationTime.Nanoseconds()
	}

	return &FileInfo{
		FileInfo:   fileStat,
		CreateTime: time.Unix(0, unixTime),
	}, nil
}
