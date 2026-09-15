package returncount

func TooMany() (int, int, int, int) { // want `function TooMany returns 4 values, limit is 3`
	return 1, 2, 3, 4
}

func AtLimit() (int, int, int) {
	return 1, 2, 3
}

func One() int {
	return 1
}

func None() {}
