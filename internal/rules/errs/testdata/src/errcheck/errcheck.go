// Package errcheck exercises unchecked error results.
package errcheck

import "os"

// Fails returns an error nobody may ignore.
func Fails() error {
	return nil
}

// Two returns a value and an error.
func Two() (int, error) {
	return 1, nil
}

// Ignored drops errors on the floor.
func Ignored() {
	Fails()     // want `error result of Fails is unchecked`
	_ = Fails() // want `error result of Fails is discarded`

	value, _ := Two() // want `error result of Two is discarded`
	_ = value

	os.Remove("x") // want `error result of Remove is unchecked`
}

// Checked handles what it is given.
func Checked() error {
	if err := Fails(); err != nil {
		return err
	}

	value, err := Two()
	if err != nil {
		return err
	}

	_ = value

	return nil
}

// NoError returns nothing to check.
func NoError() int {
	return 1
}

// Fine calls something that cannot fail.
func Fine() {
	NoError()
}
