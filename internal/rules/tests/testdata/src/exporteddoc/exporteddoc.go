// Package exporteddoc exercises documentation of the exported surface.
package exporteddoc

// Documented says what it does.
func Documented() {}

func Undocumented() {} // want `exported function Undocumented has no doc comment`

// This function does things.
func Mislabelled() {} // want `doc comment for Mislabelled must begin with its name`

func undocumentedUnexported() {}

// Thing is documented.
type Thing struct{}

type Undocumented2 struct{} // want `exported type Undocumented2 has no doc comment`
