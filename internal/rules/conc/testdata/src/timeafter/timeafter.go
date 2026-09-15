// Package timeafter exercises timers in select.
package timeafter

import "time"

// Leaks keeps a timer alive until it fires.
func Leaks(done chan int) int {
	select {
	case value := <-done:
		return value
	case <-time.After(time.Second): // want `time.After in a select leaks its timer; use time.NewTimer`
		return 0
	}
}

// Cleans up after itself.
func Cleans(done chan int) int {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	select {
	case value := <-done:
		return value
	case <-timer.C:
		return 0
	}
}
