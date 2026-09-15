package parallel

import "testing"

func helper(t *testing.T) {
	t.Helper()
	t.Parallel()
}

func TestMissing(t *testing.T) { // want `test TestMissing must call t.Parallel`
	_ = 1
}

func TestDirect(t *testing.T) {
	t.Parallel()
}

func TestDelegates(t *testing.T) {
	helper(t)
}
