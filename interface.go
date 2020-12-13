package rotatorr

//go:generate mockgen -destination=mocks/rotatorr.go -package=mocks golift.io/rotatorr Rotatorr

// Rotatorr allows passing in your own logic for file rotation.
// A couple working Rotatorr's are included with this library.
// Use those directly, or extend them with your own methods and interface.
type Rotatorr interface {
	// Rotate is called any time a file needs to be rotated.
	Rotate(fileName string) (newFile string, err error)
	// Post is called after rotation finishes and the new file is created/opened.
	// This is blocking, so if it does something like compress the rotated file,
	// it should run in a go routine and return immediately.
	Post(fileName, newFile string)

	// Dirs is called once on startup.
	// This should do any validation and return a list of directories to create.
	Dirs(fileName string) (dirPaths []string, err error)
}
