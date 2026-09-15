package noelse

func BadElse(n int) int {
	if n > 0 {
		return n
	} else { // want `else is unnecessary because the if branch ends in return`
		return -n
	}
}

func BadElseIf(n int) int {
	if n > 0 {
		return n
	} else if n < 0 { // want `else is unnecessary because the if branch ends in return`
		return -n
	}

	return 0
}

func GoodEarlyReturn(n int) int {
	if n > 0 {
		return n
	}

	return -n
}

func GoodElseWithoutReturn(n int) int {
	total := 0
	if n > 0 {
		total = n
	} else {
		total = -n
	}

	return total
}
