// Package nopanic exercises the ban on abrupt termination.
package nopanic

import (
	"os"
)

// Panics ends the program the rude way.
func Panics() {
	panic("no") // want `panic is banned; return an error instead`
}

// Exits ends the program a different rude way.
func Exits() {
	os.Exit(1) // want `os.Exit is banned outside main; return an error instead`
}

// Fine returns an error like a civilized function.
func Fine() error {
	return nil
}
