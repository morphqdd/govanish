// Package comments documents itself, which is allowed.
package comments

// Exported is documented, which is allowed.
func Exported() {
	_ = 1 // want `comment is not a doc comment`
}

// Inner has a comment inside its body, which is not allowed.
func Inner() {
	// want `comment is not a doc comment`
	_ = 1
}

// Marked carries a banned marker.
// TODO: finish this. // want `TODO and FIXME comments are banned`
func Marked() {}

//go:noinline
func Directive() {}

// Thing is documented.
type Thing struct {
	// Field is documented, which is allowed.
	Field int
}
