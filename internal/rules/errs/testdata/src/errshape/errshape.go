// Package errshape exercises where the error goes in a signature.
package errshape

// Last returns the error last, as it must.
func Last() (int, error) {
	return 1, nil
}

// First returns the error first.
func First() (error, int) { // want `error must be the last result`
	return nil, 1
}

// Named names its results correctly.
func Named() (count int, err error) {
	return 1, nil
}

// Misnamed calls the error something else.
func Misnamed() (count int, failure error) { // want `named error result must be called err, not failure`
	return 1, nil
}

// None returns nothing.
func None() {}
