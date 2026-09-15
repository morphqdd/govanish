// Package wrap exercises error construction.
package wrap

import (
	"errors"
	"fmt"
)

// Fails returns an error.
func Fails() error {
	return nil
}

// Wrapped does it right.
func Wrapped() error {
	if err := Fails(); err != nil {
		return fmt.Errorf("do the thing: %w", err)
	}

	return nil
}

// Formatted loses the cause.
func Formatted() error {
	if err := Fails(); err != nil {
		return fmt.Errorf("do the thing: %v", err) // want `wrap the error with %w, not %v`
	}

	return nil
}

// Shouted starts with a capital.
func Shouted() error {
	return errors.New("Do the thing") // want `error message must start lowercase`
}

// Punctuated ends with punctuation.
func Punctuated() error {
	return errors.New("do the thing.") // want `error message must not end with punctuation`
}

// Fine is fine.
func Fine() error {
	return errors.New("do the thing")
}
