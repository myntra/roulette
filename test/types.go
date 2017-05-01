package test

import (
	"io/ioutil"
	"log"

	"github.com/myntra/roulette"
)

// SimpleParseExpect ...
type SimpleParseExpect interface {
	Expected() bool
	Parse(fn func(val interface{}))
	Execute()
}

func readFile(path string) []byte {
	ruleFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ruleFile read err #%v  at path %v", err, path)
	}

	return ruleFile
}

// T common test object
type T struct {
	A        int
	B        int
	XML      string
	Parser   *roulette.TextTemplateParser
	Executor *roulette.SimpleExecutor

	Name        string
	Description string
}

//SetA ...
func (t *T) SetA(a int, prevVal ...bool) bool {

	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}
	t.A = a
	return true

}

// T1 ...
type T1 struct {
	T
	ExpectFunc func(t *T1) bool
}

// Execute ...
func (t *T1) Execute() {
	t.Executor.Execute(t)
}

// Parse ...
func (t *T1) Parse(fn func(val interface{})) {
	parser, err := roulette.NewSimpleParser(readFile(t.XML))
	if err != nil {
		log.Fatal(err)
	}
	t.Parser = parser.(*roulette.TextTemplateParser)
	t.Executor = roulette.NewSimpleExecutor(t.Parser).(*roulette.SimpleExecutor)
}

// Expected ...
func (t *T1) Expected() bool {
	return t.ExpectFunc(t)
}

// T2 ...
type T2 struct {
	T
	ExpectFunc func(t *T2) bool
}

// Execute ...
func (t *T2) Execute() {
	t.Executor.Execute(t)
}

// Parse ...
func (t *T2) Parse(fn func(val interface{})) {
	parser, err := roulette.NewSimpleParser(readFile(t.XML))
	if err != nil {
		log.Fatal(err)
	}
	t.Parser = parser.(*roulette.TextTemplateParser)
	t.Executor = roulette.NewSimpleExecutor(t.Parser).(*roulette.SimpleExecutor)

}

// Expected ...
func (t *T2) Expected() bool {
	return t.ExpectFunc(t)
}
