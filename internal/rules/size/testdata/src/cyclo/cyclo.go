package cyclo

func Complex(n int) int { // want `function Complex has cyclomatic complexity [0-9]+, limit is 8`
	total := 0
	if n > 0 && n < 10 {
		total++
	}
	if n > 1 {
		total++
	}
	if n > 2 {
		total++
	}
	for i := 0; i < n; i++ {
		total++
	}
	switch n {
	case 1:
		total++
	case 2:
		total++
	}
	if n > 3 || n < -3 {
		total++
	}

	return total
}

func AtLimit(n int) int {
	total := 0
	if n > 0 {
		total++
	}
	if n > 1 {
		total++
	}
	if n > 2 {
		total++
	}
	if n > 3 {
		total++
	}
	if n > 4 {
		total++
	}
	if n > 5 {
		total++
	}
	if n > 6 {
		total++
	}

	return total
}

func Simple() int {
	return 1
}
