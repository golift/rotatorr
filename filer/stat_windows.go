package filer

import (
	"os"
	"syscall"
	"time"
)

// Stat returns a *FileInfo struct w/ attached os.FileInfo interface.
func Stat(filename string) (*FileInfo, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		FileInfo:   fi,
		CreateTime: time.Unix(0, fi.Sys().(*syscall.Win32FileAttributeData).CreationTime.Nanoseconds()),
	}, nil
}
