package paramcount

func TooMany(a, b, c, d, e int) int { // want `function TooMany has 5 parameters, limit is 4`
	return a + b + c + d + e
}

func AtLimit(a, b, c, d int) int {
	return a + b + c + d
}

type Thing struct{}

func (t Thing) AlsoTooMany(a, b, c, d, e int) int { // want `method AlsoTooMany has 5 parameters, limit is 4`
	return a + b + c + d + e
}

func None() {}
