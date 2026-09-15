// Package declorderbad declares things in the wrong order.
package declorderbad

// Thing is declared before the constants and variables.
type Thing struct{}

// Limit is declared after a type.
const Limit = 10 // want `const declaration must come before type declarations`

// Count is declared after a type.
var Count = 0 // want `var declaration must come before type declarations`

func unexported() {}

// Exported is declared after an unexported function.
func Exported() {} // want `exported function Exported must be declared before unexported functions`
