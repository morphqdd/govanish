// Package noboolparam exercises the ban on boolean parameters.
package noboolparam

// Render takes a flag nobody can read at the call site.
func Render(verbose bool) {} // want `parameter verbose is a bool; a caller cannot read what true means`

// Mode is an enumeration that says what it means.
type Mode int

// Verbose is one such mode.
const Verbose Mode = 1

// RenderMode takes a named mode instead.
func RenderMode(mode Mode) {}

// Reports returning a bool is fine.
func Reports() bool {
	return true
}
