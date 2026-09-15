package nesting

func TooDeep(n int) int {
	if n > 0 {
		for i := 0; i < n; i++ {
			if i > 1 {
				if i > 2 { // want `block is nested 4 levels deep, limit is 3`
					return i
				}
			}
		}
	}

	return 0
}

func AtLimit(n int) int {
	if n > 0 {
		for i := 0; i < n; i++ {
			if i > 1 {
				return i
			}
		}
	}

	return 0
}
