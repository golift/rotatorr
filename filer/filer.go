// Package filer is an interface used in the rotatorr subpackages.
// You may override this to gain more control of operations in your app.
package filer

//go:generate mockgen -destination=../mocks/filer.go -package=mocks golift.io/rotatorr/filer Filer
//go:generate mockgen -destination=../mocks/fileinfo.go -package=mocks os FileInfo

import (
	"io/ioutil"
	"os"
	"time"
)

// Filer is used to override file-managing procedures.
type Filer interface {
	Remove(fileName string) error
	Rename(fileName, newPath string) error
	ReadDir(dirPath string) ([]os.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Stat(filename string) (*FileInfo, error)
}

// Default returns a Filer interface that works, using default procedures.
func Default() Filer {
	return &File{}
}

// FileInfo contains normal os.FileInfo + file creation time.
// Created by Stat(). Sorry in advance.
type FileInfo struct {
	os.FileInfo
	CreateTime time.Time
}

// File can be embedded in a custom type to provide the missing methods for the Filer interface.
type File struct{}

// Removes provides os.Remove.
func (f *File) Remove(fileName string) error {
	return os.Remove(fileName)
}

// Rename provides os.Rename.
func (f *File) Rename(fileName, newPath string) error {
	return os.Rename(fileName, newPath)
}

// ReadDir provides ioutil.ReadDir.
func (f *File) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

// MkdirAll provides os.MkdirAll.
func (f *File) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// OpenFile provides os.OpenFile.
func (f *File) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

// Rename provides custom file stats that wrap os.Stat output.
func (f *File) Stat(filename string) (*FileInfo, error) {
	return Stat(filename)
}
