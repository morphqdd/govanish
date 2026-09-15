// Package namelength exercises identifier length.
package namelength

// Thing is long enough.
type Thing struct{}

// Method uses a short receiver, which is allowed.
func (t Thing) Method() int {
	return 1
}

// Other uses the same receiver name, as it must.
func (t Thing) Other() int {
	return 2
}

// Loop names its index with one letter.
func Loop() int {
	total := 0

	for i := 0; i < 3; i++ { // want `identifier i is too short; names must be at least 3 characters`
		total += i
	}

	return total
}

// Named spells its index out.
func Named() int {
	total := 0

	for index := 0; index < 3; index++ {
		total += index
	}

	return total
}

// Allowed uses the names the whitelist permits.
func Allowed() error {
	err := fail()
	ok := err == nil
	_ = ok

	return err
}

func fail() error {
	return nil
}
