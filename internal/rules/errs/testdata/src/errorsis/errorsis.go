// Package errorsis exercises error comparison.
package errorsis

import "errors"

// ErrSentinel is returned when nothing else fits.
var ErrSentinel = errors.New("sentinel")

// Fails returns an error.
func Fails() error {
	return ErrSentinel
}

// Compares uses the wrong operator.
func Compares() bool {
	err := Fails()

	return err == ErrSentinel // want `compare errors with errors.Is, not ==`
}

// ComparesNot uses the wrong operator the other way.
func ComparesNot() bool {
	err := Fails()

	return err != ErrSentinel // want `compare errors with errors.Is, not !=`
}

// NilCheck is allowed, because nil is not an error value.
func NilCheck() bool {
	err := Fails()

	return err != nil
}

// Correct does it properly.
func Correct() bool {
	return errors.Is(Fails(), ErrSentinel)
}
