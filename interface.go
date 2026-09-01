package rotatorr

//go:generate mockgen -destination=mocks/rotatorr.go -package=mocks golift.io/rotatorr Rotatorr

// Rotatorr allows passing in your own logic for file rotation.
// A couple working Rotatorr's are included with this library.
// Use those directly, or extend them with your own methods and interface.
type Rotatorr interface {
	// Rotate is called any time a file needs to be rotated.
	Rotate(fileName string) (newFile string, err error)
	// Post runs on the logger dispatch goroutine after Rotate, and after a
	// successful Reopen once the live file is open. It blocks every Write until
	// it returns, so compression (or anything slow) should start a goroutine.
	// After Rotate, newFile is the backup path; Post still runs if openLog fails.
	// After Reopen, newFile is empty: rotatorr created no backup, but an external
	// tool may have moved the live file. Reopen does not call Post if openLog fails.
	Post(fileName, newFile string)

	// Dirs is called once on startup.
	// This should do any validation and return a list of directories to create.
	Dirs(fileName string) (dirPaths []string, err error)
}
