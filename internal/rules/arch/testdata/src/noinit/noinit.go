package noinit

import "os"

func init() { // want `func init is banned`
	_ = os.Getenv("HOME")
}

func NewThing() *Thing {
	return &Thing{}
}

type Thing struct{}

func (t *Thing) init() {}
