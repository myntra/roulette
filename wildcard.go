package roulette

// modified from https://siongui.github.io/2017/04/11/go-wildcard-pattern-matching/
func toRuneSlice(s string) []rune {
	var r []rune
	for _, runeVal := range s {
		r = append(r, runeVal)
	}
	return r
}

// Function that matches input str with given wildcard pattern
func wildcardMatcher(str, pattern string) bool {
	s := toRuneSlice(str)
	p := toRuneSlice(pattern)

	if len(p) == 0 {
		return len(s) == 0
	}

	matches := make([][]bool, len(s)+1)

	for i := range matches {
		matches[i] = make([]bool, len(p)+1)
	}

	matches[0][0] = true

	for j := 1; j < len(p)+1; j++ {
		if p[j-1] == '*' {
			matches[0][j] = matches[0][j-1]
		}
	}

	for i := 1; i < len(s)+1; i++ {
		for j := 1; j < len(p)+1; j++ {
			if p[j-1] == '*' {
				matches[i][j] = matches[i][j-1] || matches[i-1][j]

			} else if p[j-1] == '?' || s[i-1] == p[j-1] {
				matches[i][j] = matches[i-1][j-1]

			} else {
				matches[i][j] = false
			}
		}
	}

	return matches[len(s)][len(p)]
}
