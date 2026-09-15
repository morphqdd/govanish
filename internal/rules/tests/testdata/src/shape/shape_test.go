package shape

import "testing"

func TestGood(t *testing.T) {
	t.Parallel()
}

func Testlowercase(t *testing.T) { // want `test name Testlowercase must continue with a capital letter`
	t.Parallel()
}

func TestTableWithoutSubtests(t *testing.T) { // want `table test TestTableWithoutSubtests must run its cases with t.Run`
	t.Parallel()

	for _, name := range []string{"a", "b"} {
		_ = name
	}
}

func cases() []string {
	return []string{"a", "b"}
}

func TestRangeOverCallIsNotATable(t *testing.T) {
	t.Parallel()

	for _, name := range cases() {
		_ = name
	}
}

func TestTableInVariable(t *testing.T) { // want `table test TestTableInVariable must run its cases with t.Run`
	t.Parallel()

	names := []string{"a", "b"}
	for _, name := range names {
		_ = name
	}
}

func TestTableWithSubtests(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"a", "b"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
