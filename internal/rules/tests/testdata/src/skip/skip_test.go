package skip

import "testing"

func TestSkipped(t *testing.T) {
	t.Parallel()
	t.Skip("not today") // want `t.Skip disables a test without removing it`
}

func TestSkipNow(t *testing.T) {
	t.Parallel()
	t.SkipNow() // want `t.SkipNow disables a test without removing it`
}

func TestFine(t *testing.T) {
	t.Parallel()
}
