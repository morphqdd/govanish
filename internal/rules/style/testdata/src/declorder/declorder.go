// Package declorder exercises declaration ordering.
package declorder

// Limit comes first, as constants must.
const Limit = 10

// Count comes next, as variables must.
var Count = 0

// Thing comes after variables.
type Thing struct{}

// Exported comes before unexported functions.
func Exported() {}

func unexported() {}
