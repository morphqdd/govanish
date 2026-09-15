// Package deferunlock exercises mutex discipline.
package deferunlock

import "sync"

// Counter guards its state.
type Counter struct {
	mu    sync.Mutex
	count int
}

// Good unlocks immediately by defer.
func (c *Counter) Good() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.count
}

// Bad does work before arranging to unlock.
func (c *Counter) Bad() int {
	c.mu.Lock() // want `Lock must be followed immediately by defer Unlock`
	value := c.count
	c.mu.Unlock()

	return value
}
