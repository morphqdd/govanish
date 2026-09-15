package shape

import "testing"

type testCase struct {
	name  string
	value int
}

func cases() []testCase {
	return []testCase{{name: "a", value: 1}}
}

func TestGood(t *testing.T) {
	t.Parallel()
}

func Testlowercase(t *testing.T) { // want `test name Testlowercase must continue with a capital letter`
	t.Parallel()
}

func TestTableWithoutSubtests(t *testing.T) { // want `table test TestTableWithoutSubtests must run its cases with t.Run`
	t.Parallel()

	for _, each := range []testCase{{name: "a", value: 1}} {
		_ = each
	}
}

func TestTableInVariable(t *testing.T) { // want `table test TestTableInVariable must run its cases with t.Run`
	t.Parallel()

	table := []testCase{{name: "a", value: 1}}
	for _, each := range table {
		_ = each
	}
}

func TestRangeOverCallIsNotATable(t *testing.T) {
	t.Parallel()

	for _, each := range cases() {
		_ = each
	}
}

func TestRangeOverExpectationsIsNotATable(t *testing.T) {
	t.Parallel()

	wanted := map[string]bool{"a": true}
	for name := range wanted {
		_ = name
	}
}

func TestTableWithSubtests(t *testing.T) {
	t.Parallel()

	for _, each := range []testCase{{name: "a", value: 1}} {
		t.Run(each.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
