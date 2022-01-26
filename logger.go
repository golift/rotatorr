package rotatorr

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"golift.io/rotatorr/filer"
)

// These are the default directory and log file POSIX modes.
const (
	FileMode os.FileMode = 0600
	DirMode  os.FileMode = 0750
)

// DefaultMaxSize is only used when Every and FileSize Config
// struct members are omitted.
const DefaultMaxSize = 10 * 1024 * 1024

// Custom errors returned by this package.
var (
	ErrWriteTooLarge = fmt.Errorf("log msg length exceeds max file size")
	ErrNilInterface  = fmt.Errorf("nil Rotatorr interface provided")
)

// Config is the data needed to create a new Log Rotatorr.
type Config struct {
	Rotatorr Rotatorr      // REQUIRED: Custom log Rotatorr. Use your own or one of the provided interfaces.
	Filepath string        // Full path to log file. Set this, the default is lousy.
	FileMode os.FileMode   // POSIX mode for new files.
	DirMode  os.FileMode   // POSIX mode for new folders.
	Every    time.Duration // Maximum log file age. Rotate every hour or day, etc.
	FileSize int64         // Maximum log file size in bytes. Default is unlimited (no rotation).
}

// Logger is what you get in return for providing a Config. Use this to set log output.
// You must obtain a Logger by calling one of the New() procedures.
type Logger struct {
	config      *Config       // incoming configurtation.
	log         chan []byte   // incoming log messages passed across go routines.
	resp        chan *resp    // response sent back across go routines.
	signal      chan struct{} // used for Rotate and Close ops.
	size        int64         // the size of the active open file.
	created     time.Time     // the date the active open file was created.
	File        *os.File      // The active open file. Useful for direct writing.
	Interface   Rotatorr      // copied from config for brevity.
	filer.Filer               // overridable file system procedures.
}

// resp is used to send responses back across our go routines.
type resp struct {
	size int64
	err  error
}

// New takes in your configuration and returns a Logger you can use with
// log.SetOutput(). The provided logger handles log rotation and dispatching
// post-actions like compression.
func New(config *Config) (*Logger, error) {
	l := &Logger{config: config, Interface: config.Rotatorr, Filer: filer.Default()}
	if err := l.initialize(false); err != nil {
		return nil, err
	}

	return l, nil
}

// NewMust takes in your configuration and returns a Logger you can use with
// log.SetOutput(). If an error occurs opening the log file, making log directories,
// or rotating files it is ignored (and retried later). Do not pass a Nil Rotatorr.
func NewMust(config *Config) *Logger {
	l := &Logger{config: config, Interface: config.Rotatorr, Filer: filer.Default()}

	if err := l.initialize(true); errors.Is(err, ErrNilInterface) {
		panic(err)
	}

	return l
}

// initialize runs all the startup routines.
func (l *Logger) initialize(ignoreErrors bool) (err error) {
	defer func() {
		if err == nil || ignoreErrors {
			l.log = make(chan []byte)
			l.resp = make(chan *resp)
			l.signal = make(chan struct{})

			go l.processLogChannel()
		}
	}()

	if l.Interface == nil {
		return ErrNilInterface
	} else if err := l.setConfigDefaults(); err != nil {
		return err
	}

	err = l.checkAndRotate(0)

	return err
}

// setConfigDefaults does exactly what it says. Sets missing values.
func (l *Logger) setConfigDefaults() error {
	if l.config.Filepath == "" {
		l.config.Filepath = filepath.Join(os.TempDir(),
			filepath.Base(os.Args[0])+"-"+path.Dir(reflect.TypeOf(Logger{}).PkgPath())+".log")
	}

	if l.config.Every == 0 && l.config.FileSize == 0 {
		l.config.FileSize = DefaultMaxSize
	}

	if l.config.DirMode == 0 {
		l.config.DirMode = DirMode
	}

	if l.config.FileMode == 0 {
		l.config.FileMode = FileMode
	}

	dirs, err := l.Interface.Dirs(l.config.Filepath)
	if err != nil {
		return fmt.Errorf("validating Rotatorr: %w", err)
	}

	for _, dir := range dirs {
		if err := l.MkdirAll(dir, l.config.DirMode); err != nil {
			return fmt.Errorf("making directories for logfiles: %w", err)
		}
	}

	return nil
}

// processLogChannel runs in a go routine and reads the incoming logs channel.
// Received logs are dispatched to the write method. Replies are then sent to the
// response channel. This also handles log rotation and routine shutdown. Everything
// except specific background actions (compression?) happen in this one go routine.
func (l *Logger) processLogChannel() {
	for {
		select {
		case b := <-l.log:
			size, err := l.write(b)
			l.resp <- &resp{int64(size), err}
		case _, ok := <-l.signal:
			if !ok {
				l.signal = nil
				l.resp <- &resp{err: l.stop()}

				return
			}

			size, err := l.rotate()
			l.resp <- &resp{size, err}
		}
	}
}

// openLog opens the log file for writing.
// If the file exists, it is appended to. If it does not exist, it is created.
// Any necessary folders are also created.
func (l *Logger) openLog() error {
	err := l.MkdirAll(filepath.Dir(l.config.Filepath), l.config.DirMode)
	if err != nil {
		return fmt.Errorf("making directories for logfiles: %w", err)
	}

	perm := os.O_WRONLY | os.O_APPEND

	if info, err := l.Stat(l.config.Filepath); err != nil {
		// File doesn't exist, or something wrong, truncate it!
		perm = os.O_WRONLY | os.O_TRUNC | os.O_CREATE
		l.size = 0
		l.created = time.Now()
	} else {
		// File exists, append to it!
		l.size = info.Size()
		l.created = info.CreateTime
	}

	l.File, err = l.OpenFile(l.config.Filepath, perm, l.config.FileMode)
	if err != nil {
		return fmt.Errorf("error with new logfile: %w", err)
	}

	return nil
}

// Write sends data directly to the file. This satisfies the io.ReadCloser interface.
// You should generally not call this and instead pass *Logger into log.SetOutput().
func (l *Logger) Write(b []byte) (int, error) {
	l.log <- b
	resp := <-l.resp

	return int(resp.size), resp.err
}

// write sends a message into the log file after everyhing checks out - from a channel message.
func (l *Logger) write(b []byte) (int, error) {
	if err := l.checkAndRotate(int64(len(b))); err != nil {
		return 0, err
	}

	size, err := l.File.Write(b)
	l.size += int64(size)

	if err != nil {
		return size, fmt.Errorf("error writing log msg: %w", err)
	}

	return size, nil
}

// checkAndRotate gets the current file's size and creation time.
// Checks if it's too large or too old, and rotates it if so.
// Makes sure the log file is open and ready for writing.
func (l *Logger) checkAndRotate(size int64) error {
	if l.File == nil {
		if err := l.openLog(); err != nil {
			return err
		}
	}

	if l.config.FileSize > 0 && size > l.config.FileSize {
		return fmt.Errorf("%w: %d>%d", ErrWriteTooLarge, size, l.config.FileSize)
	}

	if (l.config.FileSize != 0 && l.size+size > l.config.FileSize) ||
		(l.config.Every != 0 && time.Now().After(l.created.Add(l.config.Every))) {
		if _, err := l.rotate(); err != nil {
			return err
		}
	}

	return nil
}

// Rotate forces the log to rotate immediately. Returns the size of the rotated log.
func (l *Logger) Rotate() (int64, error) {
	l.signal <- struct{}{}
	resp := <-l.resp

	return resp.size, resp.err
}

// rotate renames the log - from a channel message.
func (l *Logger) rotate() (int64, error) {
	size := l.size

	if err := l.close(); err != nil {
		return size, err
	}

	fpath, err := l.Interface.Rotate(l.config.Filepath)
	if fpath != "" {
		defer l.Interface.Post(l.config.Filepath, fpath)
	}

	if err != nil {
		return size, fmt.Errorf("error rotatorring: %w", err)
	}

	return size, l.openLog()
}

// Close stops the go routines, closes the active log file session and all channels.
// If another Write() is sent, a panic will ensue.
func (l *Logger) Close() error {
	defer close(l.resp)
	close(l.signal)

	return (<-l.resp).err
}

// close closes the active log file - from a channel message.
func (l *Logger) close() error {
	if l.File == nil {
		return nil
	}

	err := l.File.Close()
	l.File = nil

	if err != nil {
		return fmt.Errorf("closing log file %s: %w", l.config.Filepath, err)
	}

	return nil
}

// stop closes everything down.
func (l *Logger) stop() error {
	if l.log != nil {
		close(l.log)
	}

	l.log = nil

	return l.close()
}

// Our interface must satify an io.WriteCloser.
var _ io.WriteCloser = (*Logger)(nil)
