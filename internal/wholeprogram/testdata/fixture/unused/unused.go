// Package unused consumes one thing from the used package.
package unused

import "fixture/used"

// Call consumes Consumed, and nothing else.
func Call() int {
	return used.Consumed()
}
