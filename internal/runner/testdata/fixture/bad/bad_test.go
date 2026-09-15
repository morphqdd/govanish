package bad

import "testing"

func TestLoaded(t *testing.T) {
	if !Loaded() {
		t.Error("want loaded")
	}
}
