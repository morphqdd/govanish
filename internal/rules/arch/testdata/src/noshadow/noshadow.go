// Package noshadow exercises variable shadowing.
package noshadow

// Fails returns an error.
func Fails() error {
	return nil
}

// Shadows hides the outer err behind an inner one.
func Shadows() error {
	err := Fails()

	if err != nil {
		err := Fails() // want `declaration of err shadows an outer variable`
		_ = err
	}

	return err
}

// Distinct names things differently.
func Distinct() error {
	err := Fails()

	if err != nil {
		inner := Fails()
		_ = inner
	}

	return err
}

// Sequential reuses the same scope, which is assignment, not shadowing.
func Sequential() error {
	err := Fails()
	if err != nil {
		return err
	}

	err = Fails()

	return err
}

// Closure names its parameter after the variable it replaces, which is
// how Go avoids capturing the outer one.
func Closure() {
	err := Fails()

	apply(func(err error) {
		_ = err
	})

	_ = err
}

func apply(fn func(error)) {
	fn(nil)
}
