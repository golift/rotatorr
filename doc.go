// Package rotatorr is a log rotation module designed to plug directly into a
// standard go logger. It provides an input interface to handle log rotation,
// naming, and compression. Three additional modules packages are included to
// facilitate backup-file naming in different formats, and log compression.
//
// The New() methods return a simple io.WriteCloser that works with most log packages.
// This interface handles log rotation while providing many features and
// overridable interfaces to customize the rotation experience. Inspired by
// Lumberjack: https://github.com/natefinch/lumberjack.
//
// Use this package if you write your own log file, and you're tired of your
// log file growing indefinitely.
// The included `inrotatorr`
// and `timerotator`
// modules allow a variety of naming conventions for backup files. They also
// include options to delete old files based on age, count, or both.
//
//   https://pkg.go.dev/golift.io/rotatorr/introtator
//   https://pkg.go.dev/golift.io/rotatorr/timerotator
//
package rotatorr
