// Package main is a simple example app to write logs to see log rotation in action.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"golift.io/rotatorr"
	"golift.io/rotatorr/compressor"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/timerotator"
)

// ///////////////////////////////////////////////////////////////////////// //

/* This is a simple example app to write logs to see log rotation in action. */

// Usage, timerotator and compression:
//   go run ./cmd/exampleapp time compress
//
// Usage, introtator no compression:
//   go run ./cmd/exampleapp int
//
// Usage, introtator, rotate every everyInterval:
//   go run ./cmd/exampleapp every

const (
	logFileSize     = 1024 * 1024 // 1 megabyte.
	logFilePath     = "/tmp/myfolder/myfile.log"
	bytesPerLogLine = 5000
	timeBetweenLogs = time.Millisecond * 5
	everyInterval   = 2 * time.Second
	fileCount       = 10
)

// ///////////////////////////////////////////////////////////////////////// //

func main() {
	var (
		logger io.WriteCloser
		err    error
	)

	switch {
	case isArg("every"):
		logger, err = everyTestorr()
	case isArg("time"):
		logger, err = timeTestorr()
	case isArg("int"):
		logger, err = intTestorr()
	default:
		fmt.Println("pass test arg: time or int")
		os.Exit(1)
	}

	if err != nil {
		panic(err)
	}

	log.SetFlags(log.LstdFlags)
	log.SetOutput(logger)
	makeLogs()
}

// Write fake logs!
func makeLogs() {
	logLine := string(bytes.Repeat([]byte{'_'}, bytesPerLogLine))

	ticker := time.NewTicker(timeBetweenLogs)
	for range ticker.C {
		fmt.Print(".")

		err := log.Output(0, logLine)
		if err != nil {
			panic(err)
		}
	}
}

func timeTestorr() (io.WriteCloser, error) {
	return rotatorr.New(&rotatorr.Config{
		Filepath: logFilePath,
		FileSize: logFileSize,
		Rotatorr: &timerotator.Layout{
			FileCount:  fileCount,
			PostRotate: getPost(),
		},
	})
}

func intTestorr() (io.WriteCloser, error) {
	return rotatorr.New(&rotatorr.Config{
		Filepath: logFilePath,
		FileSize: logFileSize, // 1 megabytes.
		Rotatorr: &introtator.Layout{
			FileCount:  fileCount,
			FileOrder:  introtator.Ascending,
			PostRotate: getPost(),
		},
	})
}

func everyTestorr() (io.WriteCloser, error) {
	return rotatorr.New(&rotatorr.Config{
		Filepath: logFilePath,
		Every:    everyInterval,
		Rotatorr: &introtator.Layout{
			FileCount:  fileCount,
			FileOrder:  introtator.Ascending,
			PostRotate: getPost(),
		},
	})
}

func getPost() func(string, string) {
	if isArg("compress") {
		return func(fileName, newFile string) {
			fmt.Printf("\nfile rotated: %s -> %s\n", fileName, newFile)
			compressor.CompressBackgroundWithLog(newFile, func(s string, v ...any) { fmt.Printf(s, v...) })
			fmt.Println("compressed", newFile)
		}
	}

	return func(fileName, newFile string) {
		fmt.Printf("\nfile rotated: %s -> %s\n", fileName, newFile)
	}
}

// seems easy, but flag is better.
func isArg(arg string) bool {
	for _, a := range os.Args {
		if strings.EqualFold(a, arg) {
			return true
		}
	}

	return false
}
