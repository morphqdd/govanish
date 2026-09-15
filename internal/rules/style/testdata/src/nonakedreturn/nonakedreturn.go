package nonakedreturn

func Naked() (count int, err error) {
	count = 1

	return // want `naked return hides which values are returned`
}

func Explicit() (count int, err error) {
	return 1, nil
}

func Unnamed() (int, error) {
	return 1, nil
}

func NoResults() {
	return
}
