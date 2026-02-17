// Package compressor provides a simple interface used for
// a post-rotate Rotatorr hook that compresses files.
package compressor

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"time"

	"golift.io/rotatorr/filer"
)

// SuffixGZ is appended to a fileName to make the new compressed file name.
const SuffixGZ = ".gz"

// CompressLevel sets the global compression level.
var CompressLevel = gzip.DefaultCompression //nolint:gochecknoglobals

// Filer allows overriding os-file procedures.
var Filer = filer.Default() //nolint:gochecknoglobals

// Report contains a report of the compression operation.
// Always check for Error to make sure the New* data is valid.
type Report struct {
	OldFile string
	NewFile string
	OldSize int64
	NewSize int64
	Elapsed time.Duration
	Error   error
}

// Compress gzips a file and returns a report. Blocks until finished.
func Compress(fileName string) (*Report, error) {
	// fmt.Println("compressing", fileName)
	report := &Report{
		OldFile: fileName,
		NewFile: fileName + SuffixGZ,
		OldSize: 0,
		NewSize: 0,
		Error:   nil,
		Elapsed: 0,
	}

	level := CompressLevel
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	oldFile, err := Filer.Stat(report.OldFile)
	if report.Error = err; report.Error != nil {
		return report, fmt.Errorf("stating old file: %w", report.Error)
	}

	report.OldSize = oldFile.Size()
	start := time.Now()
	report.NewSize, report.Error = compress(report.OldFile, report.NewFile, oldFile.Mode(), level)
	report.Elapsed = time.Since(start)

	if report.Error != nil {
		return report, fmt.Errorf("compressor error: %w", report.Error)
	}

	return report, nil
}

// CompressBackground runs a file compression in the background.
// A report is sent to a provided callback function when compression finishes.
// Avoid using this on files that may be renamed by another thread.
func CompressBackground(fileName string, cb func(report *Report)) {
	go func() {
		report, _ := Compress(fileName)

		if cb != nil {
			cb(report)
		}
	}()
}

// CompressWithLog is the same as Compress, except it writes a report log instead of returning it.
func CompressWithLog(fileName string, printf func(msg string, fmt ...any)) {
	report, _ := Compress(fileName)
	go Log(report, printf) // in a go routine to avoid possible deadlock with rotatorr.
}

// CompressBackgroundPostRotate satisfies the post-rotate interface in rotatorr.
// This rotates a file and writes the success to the old log file, or
// the error to the existing log file (using the global logger).
// This is safe for use with the timerotator package.
func CompressBackgroundPostRotate(_, fileName string) {
	CompressBackgroundWithLog(fileName, nil)
}

// CompressPostRotate satisfies the post-rotate interface in rotatorr.
// This rotates a file and writes the success to the old log file, or
// the error to the existing log file (using the global logger).
// This is safe for use with the introtator package.
func CompressPostRotate(_, fileName string) {
	CompressWithLog(fileName, nil)
}

// CompressBackgroundWithLog like CompressBackground runs a file compression in
// the background, but writes a log message when finished instead of a callback.
// Avoid using this on files that may be renamed by another thread.
func CompressBackgroundWithLog(fileName string, printf func(msg string, fmt ...any)) {
	CompressBackground(fileName, func(report *Report) { Log(report, printf) })
}

// Log sends a report to a custom procedure.
func Log(report *Report, printf func(msg string, fmt ...any)) {
	if printf == nil {
		printf = log.Printf
	}

	const kilobyte = 1024

	if report.Error != nil {
		printf("Compression Error after %v: %v", report.Elapsed.Round(time.Second), report.Error)
	} else {
		printf("Compression Finished in %v: %s/%dkB -> %s/%dkB", report.Elapsed.Round(time.Second),
			report.OldFile, report.OldSize/kilobyte, report.NewFile, report.NewSize/kilobyte)
	}
}

// compress does the "hard" work: Open the old file, open the new file, create a gzip writer,
// copy the writer to the new file, close all open file handles, and lastly delete the old file.
func compress(oldFile, newFile string, mode os.FileMode, level int) (int64, error) {
	var (
		size     int64
		err      error
		ncf, gzf *os.File
	)

	defer func() { // First, so it executes last.
		if err != nil {
			_ = Filer.Remove(newFile)
		} else {
			_ = Filer.Remove(oldFile)
		}
	}()

	ncf, err = Filer.OpenFile(oldFile, os.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("opening source file: %w", err)
	}
	defer ncf.Close()

	gzf, err = Filer.OpenFile(newFile, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return 0, fmt.Errorf("opening gz file: %w", err)
	}

	defer func() {
		gzf.Close()
		// Set size of new file.
		if fs, err := Filer.Stat(newFile); err == nil {
			size = fs.Size()
		}
	}()

	gzw, _ := gzip.NewWriterLevel(gzf, level)
	defer gzw.Close()
	gzw.Comment = reflect.TypeFor[Report]().PkgPath()

	size, err = io.Copy(gzw, ncf)
	if err != nil {
		return size, fmt.Errorf("%s -> %s: %w", oldFile, newFile, err)
	}

	return size, nil
}
