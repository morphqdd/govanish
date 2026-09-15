// Package nobareprim exercises bare primitives in exported signatures.
package nobareprim

// UserID is a domain type.
type UserID string

// Find takes a domain type, as it must.
func Find(id UserID) *User {
	return nil
}

// User is a domain type.
type User struct{}

// Lookup takes a bare string, which says nothing about what it holds.
func Lookup(id string) *User { // want `parameter id is a bare string; define a domain type`
	return nil
}

// Count returns a bare int.
func Count() int { // want `result of Count is a bare int; define a domain type`
	return 0
}

// Ok returns a bool, which needs no domain type.
func Ok() bool {
	return true
}

// Fails returns an error, which needs no domain type.
func Fails() error {
	return nil
}

func lookup(id string) *User {
	return nil
}
