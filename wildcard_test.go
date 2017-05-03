package roulette

import "testing"

func TestWildcardMatcher(t *testing.T) {
	if wildcardMatcher("", "") != true {
		t.Error("err")
	}
	if wildcardMatcher("a", "") != false {
		t.Error("err")
	}
	if wildcardMatcher("baaabab", "*****ba*****ab") != true {
		t.Error("err")
	}
	if wildcardMatcher("baaabab", "baaa?ab") != true {
		t.Error("err")
	}
	if wildcardMatcher("baaabab", "ba*a?") != true {
		t.Error("err")
	}
	if wildcardMatcher("baaabab", "a*ab") != false {
		t.Error("err")
	}
	if wildcardMatcher("aa", "a") != false {
		t.Error("err")
	}
	if wildcardMatcher("aa", "aa") != true {
		t.Error("err")
	}
	if wildcardMatcher("aaa", "aa") != false {
		t.Error("err")
	}
	if wildcardMatcher("aa", "*") != true {
		t.Error("err")
	}
	if wildcardMatcher("aa", "a*") != true {
		t.Error("err")
	}
	if wildcardMatcher("ab", "?*") != true {
		t.Error("err")
	}
	if wildcardMatcher("aab", "c*a*b") != false {
		t.Error("err")
	}
}
