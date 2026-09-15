// Package keyedliterals exercises composite literal field names.
package keyedliterals

// Point is a struct with two indistinguishable fields.
type Point struct {
	X int
	Y int
}

// Unkeyed relies on field order.
func Unkeyed() Point {
	return Point{1, 2} // want `composite literal of Point must name its fields`
}

// Keyed says which is which.
func Keyed() Point {
	return Point{X: 1, Y: 2}
}

// Slices are not structs, so their elements need no names.
func Slices() []int {
	return []int{1, 2, 3}
}

// Maps already have keys.
func Maps() map[string]int {
	return map[string]int{"a": 1}
}
