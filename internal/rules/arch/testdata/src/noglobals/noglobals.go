// Package noglobals exercises the ban on package-level variables.
package noglobals

// Limit is a constant, which is allowed.
const Limit = 10

// Counter is mutable package state, which is not.
var Counter = 0 // want `package-level var Counter is banned; pass state explicitly`

// Local keeps its state inside.
func Local() int {
	counter := 0

	return counter
}
