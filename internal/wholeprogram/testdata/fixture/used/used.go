// Package used exports things another package consumes.
package used

// Consumed is used from outside this package.
func Consumed() int {
	return helper()
}

// Orphan is exported but nobody outside this package calls it.
func Orphan() int {
	return 1
}

func helper() int {
	return 2
}
