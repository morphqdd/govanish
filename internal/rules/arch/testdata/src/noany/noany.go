// Package noany exercises the ban on the empty interface.
package noany

// Box holds anything, which says nothing.
type Box struct {
	Value any // want `any is banned; name the type you mean`
}

// Take accepts anything.
func Take(value interface{}) {} // want `any is banned; name the type you mean`

// Give returns anything.
func Give() any { // want `any is banned; name the type you mean`
	return nil
}

// Typed says what it means.
func Typed(value int) int {
	return value
}

// Reader is a real interface, which is allowed.
type Reader interface {
	Read() int
}

type hidden struct {
	value any
}

func unexported(value any) any {
	return value
}
