
# Rotatorr

## Go App Log Rotation!

[![GoDoc](https://godoc.org/golift.io/rotatorr/svc?status.svg)](https://godoc.org/golift.io/rotatorr)
[![Go Report Card](https://goreportcard.com/badge/golift.io/rotatorr)](https://goreportcard.com/report/golift.io/rotatorr)
[![MIT License](http://img.shields.io/:license-mit-blue.svg)](https://github.com/golift/rotatorr/blob/master/LICENSE)
[![discord](https://badgen.net/badge/icon/Discord?color=0011ff&label&icon=https://simpleicons.now.sh/discord/eee "GoLift Discord")](https://golift.io/discord)

### Description

Rotatorr provides a simple `io.WriteCloser` you can plug into the default `log`
package. This interface handles log rotation while providing many features and
overridable interfaces to customize the rotation experience. Inspired by
[Lumberjack](https://github.com/natefinch/lumberjack). I wrote this because I
wanted integer log files, and I figured why not fix a few of the things
reported in the Lumberjack issues and pull requests.

### Simple Usage

This example rotates logs once they reach 10mb. The backup log files have a
time stamp written to their name.
```go
log.SetOutput(rotatorr.NewMust(&rotatorr.Config{
	Filesize: 1024 * 1024 * 10, // 10 megabytes
	Filepath: "/var/log/service.log",
	Rotatorr: &timerotator.Layout{FileCount: 10},
}))
```

### Advanced Usage

In the example above you can see that the `Rotatorr` interface is satisfied by
`*timerotator.Layout`. The other built-in option is `*introtator.Layout`.

As a version 0 package, some of the interfaces are bound to change as we find bugs
and make further improvements. Feedback and bug reports are welcomed and encouraged!

The [time rotator](https://pkg.go.dev/golift.io/rotatorr/timerotator)
puts time stamps in the backup log file names.
The [int rotator](https://pkg.go.dev/golift.io/rotatorr/introtator)
uses an integer (like `logfile.1.log`). Pick one and stick with it for best results.
You may also enable compression by adding a callback to either rotator that calls
the included [compressor](https://pkg.go.dev/golift.io/rotatorr/compressor) library.
**All the advanced examples are in [godoc](https://pkg.go.dev/golift.io/rotatorr)**,
or just check out the [examples_test.go](examples_test.go) file in this repo and the
[example app](cmd/exampleapp/main.go) that's included.
Below you'll find the three main data structures you can provide to this
package to make log file rotation work just the way you want.

#### Type: `rotatorr.Config`

All of the struct members are optional except the `Rotatorr` interface.

```go
type Config struct {
	Filepath string        // Full path to log file. Set this, the default is lousy.
	FileMode os.FileMode   // POSIX mode for new files.
	DirMode  os.FileMode   // POSIX mode for new folders.
	Every    time.Duration // Maximum log file age. Rotate every hour or day, etc.
	FileSize int64         // Maximum log file size in bytes. Default is unlimited (no rotation).
	Rotatorr Rotatorr      // REQUIRED: Custom log Rotatorr. Use your own or one of the provided interfaces.
}
```

#### Type: `introtator.Layout`

 -   `Rotatorr` interface.

```go
type Layout struct {
	ArchiveDir string // Location where rotated backup logs are moved to.
	FileCount  int    // Maximum number of rotated log files.
	FileOrder  Order  // Control the order of the integer-named backup log files.
	PostRotate func(fileName, newFile string)
}
```

#### Type: `timerotator.Layout`

-   `Rotatorr` interface.

```go
type Layout struct {
	ArchiveDir string        // Location where rotated backup logs are moved to.
	FileCount  int           // Maximum number of rotated log files.
	FileAge    time.Duration // Maximum age of rotated files.
	UseUTC     bool          // Sets the time zone to UTC when writing Time Formats (backup files).
	Format     string        // Format for Go Time. Used as the name.
	Joiner     string        // The string betwene the file name prefix and time stamp. Default: -
	PostRotate func(fileName, newFile string)
}
```
