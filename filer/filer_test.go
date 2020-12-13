package filer_test

import (
	"fmt"

	"golift.io/rotatorr/filer"
)

// Our interface must satify a filer.Filer.
var _ filer.Filer = (*MyFiler)(nil)

// Create a custom Filer that overrides only the Rename method.
type MyFiler struct {
	filer.File
}

func (f *MyFiler) Rename(oldpath, newpath string) error {
	fmt.Printf("Renamed %s -> %s\n", oldpath, newpath)

	return nil
}

func ExampleFile() {
	// Pass s into any package that uses a filer.Filer.
	s := &MyFiler{}
	_ = s.Rename("old.file", "new.file")
	// Output:
	// Renamed old.file -> new.file
}
