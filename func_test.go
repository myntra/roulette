package roulette

import (
	"bytes"
	"fmt"
	"log"
	"testing"
	"text/template"
)

type cmpTest struct {
	expr  string
	truth string
	ok    bool
}

var cmpTests = []cmpTest{
	{"eq 1 2 | in 2 1 3 ", "false", true},
	{"in 2 1 3 ", "true", true},
	{"eq 1 1 | in 2 1 3 ", "true", true},
	{"and (eq 1 1) (ne 1 2)", "true", true},
	{"eq 1 1 | and (eq 1 1) (ne 1 2)", "true", true},
	{"eq 1 2 | and (eq 1 1) (ne 1 2)", "false", true},
	{"or (eq 1 1) (ne 1 1)", "true", true},
	{"eq 1 2 | or (eq 1 1) (ne 1 1)", "false", true},
	{"ne 1 2 | or (eq 1 1) (ne 1 1)", "true", true},
	{"or (eq 2 1) (ne 1 2)", "true", true},
	{"or (eq 2 1) (ne 1 1)", "false", true},
	{"not false", "true", true},
	{"eq 1 2 | not false", "false", true},
	{"eq 1 1 | not false", "true", true},
	{"eq true true", "true", true},
	{"ne 1 2 | eq true true", "true", true},
	{"eq true false", "false", true},
	{"eq 1+2i 1+2i", "true", true},
	{"eq 1+2i 1+3i", "false", true},
	{"eq 1.5 1.5", "true", true},
	{"eq 1.5 2.5", "false", true},
	{"eq 1 1", "true", true},
	{"eq 1 2", "false", true},
	{"eq `xy` `xy`", "true", true},
	{"eq `xy` `xyz`", "false", true},
	{"eq .Uthree .Uthree", "true", true},
	{"eq .Uthree .Ufour", "false", true},
	{"eq 3 4 5 6 3", "true", true},
	{"eq 3 4 5 6 7", "false", true},
	{"ne true true", "false", true},
	{"eq 1 1 | ne true true", "false", true},
	{"ne true false", "true", true},
	{"eq 1 2 | ne true false", "false", true},
	{"ne 1+2i 1+2i", "false", true},
	{"ne 1+2i 1+3i", "true", true},
	{"ne 1.5 1.5", "false", true},
	{"ne 1.5 2.5", "true", true},
	{"ne 1 1", "false", true},
	{"ne 1 2", "true", true},
	{"ne `xy` `xy`", "false", true},
	{"ne `xy` `xyz`", "true", true},
	{"ne .Uthree .Uthree", "false", true},
	{"ne .Uthree .Ufour", "true", true},
	{"lt 1.5 1.5", "false", true},
	{"eq 1 1 | lt 1.5 1.5", "false", true},
	{"lt 1.5 2.5", "true", true},
	{"eq 1 2 | lt 1.5 2.5", "false", true},
	{"lt 1 1", "false", true},
	{"lt 1 2", "true", true},
	{"lt `xy` `xy`", "false", true},
	{"lt `xy` `xyz`", "true", true},
	{"lt .Uthree .Uthree", "false", true},
	{"lt .Uthree .Ufour", "true", true},
	{"le 1.5 1.5", "true", true},
	{"eq 1 1 | le 1.5 1.5", "true", true},
	{"le 1.5 2.5", "true", true},
	{"le 2.5 1.5", "false", true},
	{"le 1 1", "true", true},
	{"le 1 2", "true", true},
	{"le 2 1", "false", true},
	{"le `xy` `xy`", "true", true},
	{"le `xy` `xyz`", "true", true},
	{"le `xyz` `xy`", "false", true},
	{"le .Uthree .Uthree", "true", true},
	{"le .Uthree .Ufour", "true", true},
	{"le .Ufour .Uthree", "false", true},
	{"gt 1.5 1.5", "false", true},
	{"eq 1 2 | gt 1.5 1.5", "false", true},
	{"gt 1.5 2.5", "false", true},
	{"gt 1 1", "false", true},
	{"gt 2 1", "true", true},
	{"gt 1 2", "false", true},
	{"gt `xy` `xy`", "false", true},
	{"gt `xy` `xyz`", "false", true},
	{"gt .Uthree .Uthree", "false", true},
	{"gt .Uthree .Ufour", "false", true},
	{"gt .Ufour .Uthree", "true", true},
	{"ge 1.5 1.5", "true", true},
	{"eq 1 1 | ge 1.5 1.5", "true", true},
	{"ge 1.5 2.5", "false", true},
	{"ge 2.5 1.5", "true", true},
	{"ge 1 1", "true", true},
	{"ge 1 2", "false", true},
	{"ge 2 1", "true", true},
	{"ge `xy` `xy`", "true", true},
	{"ge `xy` `xyz`", "false", true},
	{"ge `xyz` `xy`", "true", true},
	{"ge .Uthree .Uthree", "true", true},
	{"ge .Uthree .Ufour", "false", true},
	{"ge .Ufour .Uthree", "true", true},
	// Mixing signed and unsigned integers.
	{"eq .Uthree .Three", "true", true},
	{"eq .Three .Uthree", "true", true},
	{"le .Uthree .Three", "true", true},
	{"le .Three .Uthree", "true", true},
	{"ge .Uthree .Three", "true", true},
	{"ge .Three .Uthree", "true", true},
	{"lt .Uthree .Three", "false", true},
	{"lt .Three .Uthree", "false", true},
	{"gt .Uthree .Three", "false", true},
	{"gt .Three .Uthree", "false", true},
	{"eq .Ufour .Three", "false", true},
	{"lt .Ufour .Three", "false", true},
	{"gt .Ufour .Three", "true", true},
	{"eq .NegOne .Uthree", "false", true},
	{"eq .Uthree .NegOne", "false", true},
	{"ne .NegOne .Uthree", "true", true},
	{"ne .Uthree .NegOne", "true", true},
	{"lt .NegOne .Uthree", "true", true},
	{"lt .Uthree .NegOne", "false", true},
	{"le .NegOne .Uthree", "true", true},
	{"le .Uthree .NegOne", "false", true},
	{"gt .NegOne .Uthree", "false", true},
	{"gt .Uthree .NegOne", "true", true},
	{"ge .NegOne .Uthree", "false", true},
	{"ge .Uthree .NegOne", "true", true},
	{"eq (index `x` 0) 'x'", "true", true}, // The example that triggered this rule.
	{"eq (index `x` 0) 'y'", "false", true},
	// Errors
	{"eq `xy` 1", "", false},    // Different types.
	{"eq 2 2.0", "", false},     // Different types.
	{"lt true true", "", false}, // Unordered types.
	{"lt 1+0i 1+0i", "", false}, // Unordered types.
}

func TestComparison(t *testing.T) {
	b := new(bytes.Buffer)
	var cmpStruct = struct {
		Uthree, Ufour uint
		NegOne, Three int
	}{3, 4, -1, 3}
	for _, test := range cmpTests {
		text := fmt.Sprintf("{{if %s}}true{{else}}false{{end}}", test.expr)
		tmpl, err := template.New("empty").Funcs(defaultFuncMap).Parse(text)
		if err != nil {
			t.Fatalf("%q: %s", test.expr, err)
		}
		b.Reset()
		err = tmpl.Execute(b, &cmpStruct)
		if test.ok && err != nil {
			t.Errorf("%s errored incorrectly: %s", test.expr, err)
			continue
		}
		if !test.ok && err == nil {
			t.Errorf("%s did not error", test.expr)
			continue
		}
		if b.String() != test.truth {
			t.Errorf("%s: want %s; got %s", test.expr, test.truth, b.String())
		}
	}
}

func TestValidateFuncs(t *testing.T) {

	testFuncMap1 := template.FuncMap{
		"test1": func(a, b int, prev ...bool) bool {
			return true
		},
	}

	err := validateFuncs(testFuncMap1)
	if err != nil {
		log.Fatal("testFuncMap1 must be valid")
	}

	testFuncMap2 := template.FuncMap{
		"test2": func(a, b int, prev ...bool) (bool, error) {
			return false, nil
		},
	}

	err = validateFuncs(testFuncMap2)
	if err != nil {
		log.Fatal("testFuncMap2 must be valid")
	}

	testFuncMap3 := template.FuncMap{
		"test3": func(a, b int, prev ...bool) (bool, int) {
			return false, 0
		},
	}

	err = validateFuncs(testFuncMap3)
	if err == nil {
		log.Fatal("testFuncMap3 must be invalid")
	}

	var _f = func(a, b int, prev ...bool) (bool, error) {
		return false, nil
	}
	testFuncMap4 := template.FuncMap{
		"_%f": _f,
	}

	err = validateFuncs(testFuncMap4)
	if err == nil {
		log.Fatal("testFuncMap4 must be invalid")
	}

	testFuncMap4 = template.FuncMap{
		"_f%f": _f,
	}

	err = validateFuncs(testFuncMap4)
	if err == nil {
		log.Fatal("testFuncMap4 must be invalid")
	}

	testFuncMap4 = template.FuncMap{
		"": _f,
	}

	err = validateFuncs(testFuncMap4)
	if err == nil {
		log.Fatal("testFuncMap4 must be invalid")
	}

	testFuncMap4 = template.FuncMap{
		"": err,
	}

	err = validateFuncs(testFuncMap4)
	if err == nil {
		log.Fatal("testFuncMap4 must be invalid")
	}

}
