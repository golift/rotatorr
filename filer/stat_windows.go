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

	creatTime, _ := fileStat.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds()

	return &FileInfo{
		FileInfo:   fileStat,
		CreateTime: time.Unix(0, creatTime),
	}, nil
}
