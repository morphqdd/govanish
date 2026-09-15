// Package sizedchan exercises channel construction.
package sizedchan

// Unbuffered hides a synchronisation decision.
func Unbuffered() chan int {
	return make(chan int) // want `make\(chan\) must state a buffer size`
}

// Buffered states its size.
func Buffered() chan int {
	return make(chan int, 1)
}

// Explicit zero is a decision, written down.
func Explicit() chan int {
	return make(chan int, 0)
}
