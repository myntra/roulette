package roulette

import (
	"io/ioutil"
	"log"
	"testing"
)

// SimpleParseExpect ...
type SimpleParseTestCase interface {
	Parse()
	Execute()
	Expected() bool
	Name() string
	Desc() string
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
	A           int
	B           int
	XML         string
	Callback    func(val interface{})
	Parser      *TextTemplateParser
	Executor    *SimpleExecutor
	ExpectFunc  func(val interface{}) bool
	TestName    string
	Description string
}

//SetA ...
func (t *T) SetA(a int, prevVal ...bool) bool {
	//	fmt.Println("SetA", a, prevVal)
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
	*T
}

// Name ...
func (t *T1) Name() string {
	return t.TestName
}

// Desc ...
func (t *T1) Desc() string {
	return t.Description
}

// Execute ...
func (t *T1) Execute() {
	t.Executor.Execute(t)
}

// Parse ...
func (t *T1) Parse() {

	var parser Parser
	var err error

	if t.Callback != nil {
		parser, err = NewCallbackParser(readFile(t.XML), t.Callback)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		parser, err = NewSimpleParser(readFile(t.XML))
		if err != nil {
			log.Fatal(err)
		}
	}

	t.Parser = parser.(*TextTemplateParser)
	t.Executor = NewSimpleExecutor(t.Parser).(*SimpleExecutor)
}

// Expected ...
func (t *T1) Expected() bool {
	return t.ExpectFunc(t)
}

var simpleParseTestCases = []SimpleParseTestCase{
	&T1{
		T: &T{
			A:           1,
			B:           2,
			XML:         "testrules/rules_simple_1.xml",
			TestName:    "TestT1SetA ",
			Description: "Expects T1.A to be 5",
			ExpectFunc: func(val interface{}) bool {
				t, ok := val.(*T1)
				//fmt.Println("ExpectFunc", val, reflect.TypeOf(val))
				if !ok {
					log.Println("expected val to be T1")
					return false
				}
				//	fmt.Println("ExpectFunc", t.A)
				if t.A != 5 {
					return false
				}
				return true
			}},
	},
}

// test xml: tags and attributes, malformed etc.
// test template: valid for types, malformed etc.
// test parsers and executors: types, nil values etc.

func TestSimpleParser(tt *testing.T) {
	for _, testcase := range simpleParseTestCases {
		testcase.Parse()
		testcase.Execute()
		if !testcase.Expected() {
			log.Fatal(testcase.Name(), testcase.Desc())
		}
	}
}
