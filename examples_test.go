package rotatorr_test

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golift.io/rotatorr"
	"golift.io/rotatorr/compressor"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/timerotator"
)

// This examples shows how to create backup logs files just like
// https://github.com/natefinch/lumberjack.
// This will rotate files at 100Mb and keep all old files (no cleanup).
// Backup log files are named with a time stamp. Compression isn't enabled.
func Example_lumberjack() {
	log.SetOutput(rotatorr.NewMust(&rotatorr.Config{
		Filepath: "/var/log/file.log", // optional.
		FileSize: 100 * 1024 * 1024,   // 100 megabytes.
		DirMode:  0o755,               // world-readable.
		Rotatorr: &timerotator.Layout{
			FileAge:   0,    // keep all backup files (default).
			FileCount: 0,    // keep all backup files (default).
			UseUTC:    true, // not the default.
		},
	}))
}

// This example demonstrates how to trigger an action after a file is rotated.
// All of the struct members for rotatorr.Config and timerotator.Layout are shown.
func Example_postRotateLog() {
	const (
		TenMB  = 10 * 1024 * 1024
		OneDay = time.Hour * 24
		Month  = time.Hour * 24 * 30
		Keep   = 10
	)

	postRotate := func(fileName, newFile string) {
		// This must run in a go routine or a deadlock will occur when calling log.Printf.
		// If you're doing things besides logging, you do not need a go routine, but this
		// function blocks logs, so make it snappy.
		go func() {
			log.Printf("file rotated: %s -> %s", fileName, newFile)
		}()
	}

	rotator, err := rotatorr.New(&rotatorr.Config{
		Filepath: "/var/log/file.log", // not required, but recommended.
		FileSize: TenMB,               // 10 megabytes.
		FileMode: rotatorr.FileMode,   // default: 0600
		DirMode:  rotatorr.DirMode,    // default: 0750
		Every:    OneDay,              // rotate every day
		Rotatorr: &timerotator.Layout{ // required.
			FileCount:  Keep,                      // keep 10 files
			FileAge:    Month,                     // delete files older than 30 days
			Format:     timerotator.FormatDefault, // This is the default Time Format.
			ArchiveDir: "/var/log/archives",       // override backup log file location.
			PostRotate: postRotate,                // optional post-rotate function.
			UseUTC:     false,                     // default is false.
			Joiner:     "-",                       // prefix and time stamp separator.
			Filer:      nil,                       // use default: os.Remove
		},
	})
	if err != nil {
		panic(err)
	}

	log.SetOutput(rotator)
}

func ExampleNew() {
	rotator, err := rotatorr.New(&rotatorr.Config{
		Filepath: "/var/log/service.log",
		Rotatorr: &introtator.Layout{FileCount: 10},
	})
	if err != nil {
		panic(err)
	}

	log.SetOutput(rotator)
}

func ExampleNewMust() {
	log.SetOutput(rotatorr.NewMust(&rotatorr.Config{
		FileSize: 1024 * 1024 * 100, // 100 megabytes
		Filepath: "/var/log/service.log",
		Rotatorr: &timerotator.Layout{FileCount: 10},
	}))
}

// Rotate a log on SIGHUP signal.
func ExampleLogger_Rotate() {
	rotator := rotatorr.NewMust(&rotatorr.Config{
		Filepath: "/var/log/service.log",
		Rotatorr: &timerotator.Layout{FileCount: 10},
	})
	log.SetOutput(rotator)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP)

	go func() {
		<-sigc

		_, err := rotator.Rotate()
		if err != nil {
			panic(err)
		}
	}()
}

// This is a simple example that enables log compression.
// Enabling compression on "Ascending Integer" log files is not recommend because
// it's possible for a log to be rotated (renamed) while being compressed.
// This is best utilized on Descending Integer or Time-based log files.
// Of course, these are all interfaces you can override, so customize away!
// The called CompressPostRotate procedure runs a compression in the background,
// and prints a log message when it completes.
func Example_compressor() {
	log.SetOutput(rotatorr.NewMust(&rotatorr.Config{
		Filepath: "/var/log/file.log",
		FileSize: 100 * 1024 * 1024, // 100 megabytes.
		Rotatorr: &timerotator.Layout{
			PostRotate: compressor.CompressPostRotate,
		},
	}))
}

// Example_compressor_log shows how to format a post-rotate compression log line.
func Example_compressorWithLog() {
	post := func(_, fileName string) {
		printf := func(_ string, v ...any) {
			log.Printf("[Rotatorr] %s", v...)
		}
		compressor.CompressBackgroundWithLog(fileName, printf)
	}

	log.SetOutput(rotatorr.NewMust(&rotatorr.Config{
		Filepath: "/var/log/file.log",
		FileSize: 10 * 1024 * 1024, // 10 megabytes.
		Rotatorr: &timerotator.Layout{PostRotate: post},
	}))
}

// Example_compressor_capture shows how to capture the response from a
// post-rotate compression so you can do whatever you want with it.
func Example_compressorCaptureOutput() {
	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: "/var/log/file.log",
		FileSize: 100 * 1024 * 1024, // 100 megabytes.
		Rotatorr: &timerotator.Layout{
			PostRotate: func(_, fileName string) {
				compressor.CompressBackground(fileName, func(report *compressor.Report) {
					if report.Error != nil {
						log.Printf("[Rotatorr] Error: %v", report.Error)
					} else {
						log.Printf("[Rotatorr] Compressed: %s -> %s", report.OldFile, report.NewFile)
					}
				})
			},
		},
	})
	if err != nil {
		panic(err)
	}

	log.SetOutput(logger)
}
