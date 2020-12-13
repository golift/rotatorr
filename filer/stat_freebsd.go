package filer

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// Stat returns a *FileInfo struct w/ attached os.FileInfo interface.
func Stat(filename string) (*FileInfo, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("stat err: %w", err)
	}

	fileinfo := fi.Sys().(*syscall.Stat_t)

	return &FileInfo{
		FileInfo:   fi,
		CreateTime: time.Unix(int64(fileinfo.Ctimespec.Sec), int64(fileinfo.Ctimespec.Nsec)), // nolint: unconvert
	}, nil
}
