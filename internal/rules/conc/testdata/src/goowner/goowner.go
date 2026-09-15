// Package goowner exercises goroutine ownership.
package goowner

import "sync"

// Orphan starts work nobody waits for.
func Orphan() {
	go func() { // want `go statement needs a sync.WaitGroup or errgroup.Group in scope to wait for it`
		_ = 1
	}()
}

// Owned waits for what it starts.
func Owned() {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = 1
	}()

	wg.Wait()
}
