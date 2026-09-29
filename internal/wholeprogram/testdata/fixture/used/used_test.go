package used

import "testing"

// TestHelper gives this package an in-package test, so loading with
// Tests set produces a second, test-augmented variant of it.
func TestHelper(t *testing.T) {
	t.Parallel()

	if helper() != 2 {
		t.Fatal("helper changed")
	}
}
