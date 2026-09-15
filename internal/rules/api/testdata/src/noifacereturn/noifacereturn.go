// Package noifacereturn exercises what an exported function may return.
package noifacereturn

// Reader is an interface this package defines.
type Reader interface {
	Read() int
}

type reader struct{}

func (r reader) Read() int { return 0 }

// NewReader returns an interface, hiding what the caller actually got.
func NewReader() Reader { // want `exported function NewReader must not return an interface`
	return reader{}
}

// NewConcrete returns something the caller can see.
func NewConcrete() *Concrete {
	return &Concrete{}
}

// Concrete is a concrete type.
type Concrete struct{}

// Fails returns an error, which is the one interface always allowed.
func Fails() error {
	return nil
}

func newUnexported() Reader {
	return reader{}
}
