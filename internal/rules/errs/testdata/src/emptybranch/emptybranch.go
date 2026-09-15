// Package emptybranch exercises empty error branches.
package emptybranch

// Fails returns an error.
func Fails() error {
	return nil
}

// Empty swallows the error.
func Empty() {
	if err := Fails(); err != nil { // want `error branch is empty, so the error is silently dropped`
	}
}

// Handled does something with it.
func Handled() error {
	err := Fails()
	if err != nil {
		return err
	}

	return nil
}
