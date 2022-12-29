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

	fileinfo, _ := fileStat.Sys().(*syscall.Stat_t)

	return &FileInfo{
		FileInfo:   fileStat,
		CreateTime: time.Unix(int64(fileinfo.Ctimespec.Sec), int64(fileinfo.Ctimespec.Nsec)), //nolint:unconvert
	}, nil
}
