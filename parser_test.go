package roulette

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"
	"text/template"
	"time"
)

// SimpleParseExpect ...
type SimpleParseTestCase interface {
	Parse()
	Execute()
	Expected() bool
	Name() string
	Desc() string
	SetLogLevel(string)
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
	Parser      TextTemplateParser
	Executor    *SimpleExecutor
	ExpectFunc  func(val interface{}) bool
	TestName    string
	Description string
	Loglevel    string
}

//SetA ...
func (t *T) SetA(a int, prevVal ...bool) bool {
	//fmt.Println("SetA", a, prevVal)
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

func (t *T1) SetLogLevel(level string) {
	t.Loglevel = level
}

// Parse ...
func (t *T1) Parse() {

	var err error

	config := TextTemplateParserConfig{
		LogLevel: t.Loglevel,
	}
	if t.Callback != nil {
		config.Result = NewResultCallback(t.Callback)
	}

	parser, err := NewParser(readFile(t.XML), config)
	if err != nil {
		log.Fatal(err)
	}

	t.Parser = parser.(TextTemplateParser)
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
			XML:         "testrules/rules_simple.xml",
			TestName:    "TestT1SetA ",
			Description: "Expects T1.A to be 5",
			ExpectFunc: func(val interface{}) bool {
				t, ok := val.(*T1)
				//	fmt.Println("ExpectFunc", val, reflect.TypeOf(val))
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

func testSimpleParser(loglevel string) {
	for _, testcase := range simpleParseTestCases {
		testcase.SetLogLevel(loglevel)
		testcase.Parse()
		testcase.Execute()
		if !testcase.Expected() {
			log.Fatal(testcase.Name(), testcase.Desc())
		}
	}
}

func TestSimpleParser(tt *testing.T) {
	testSimpleParser("info")
}

type T2 struct {
	A int
	B int
}

//SetA ...
func (t *T2) SetA(a int, prevVal ...bool) bool {
	//	fmt.Println("SetA", a, prevVal)
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}
	t.A = a
	return true

}

func TestArraySameType(t *testing.T) {
	t21 := &T2{A: 1, B: 2}
	t22 := &T2{A: 1, B: 2}

	config := TextTemplateParserConfig{}

	parser, err := NewParser(readFile("testrules/rules_array_same_type.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewSimpleExecutor(parser)
	executor.Execute(t21, t22)

	if t21.A != 5 {
		log.Fatal("Expected value to be 5")
	}
}

var workflowPatterns = []struct {
	workflowPattern string
	expectedVal     int
}{
	{"ipl*", 10},    // should match 1
	{"summer*", 20}, // should match 1
	{"", 20},        // should match all
}

func TestRulesetWorflow(t *testing.T) {

	for _, v := range workflowPatterns {
		t21 := &T2{A: 1, B: 2}
		config := TextTemplateParserConfig{
			WorkflowPattern:           v.workflowPattern,
			IsWildcardWorkflowPattern: true,
		}

		parser, err := NewParser(readFile("testrules/rules_workflows.xml"), config)
		if err != nil {
			log.Fatal(err)
		}

		executor := NewSimpleExecutor(parser)
		executor.Execute(t21)

		if t21.A != v.expectedVal {
			log.Fatalf("Expected value to be %d got %d", v.expectedVal, t21.A)
		}

	}
}

func TestRulesetPriorites(t *testing.T) {

	t21 := &T2{A: 1, B: 2}
	config := TextTemplateParserConfig{
		WorkflowPattern:           "ruleset1",
		IsWildcardWorkflowPattern: true,
	}

	parser, err := NewParser(readFile("testrules/rules_priorities.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewSimpleExecutor(parser)
	executor.Execute(t21)

	if t21.A != 10 {
		log.Fatalf("Expected value to 10, is %d", t21.A)
	}

	config = TextTemplateParserConfig{
		WorkflowPattern:           "ruleset2",
		IsWildcardWorkflowPattern: true,
	}

	t22 := &T2{A: 1, B: 2}

	parser, err = NewParser(readFile("testrules/rules_priorities.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor = NewSimpleExecutor(parser)
	executor.Execute(t22)

	if t22.A != 5 {
		log.Fatalf("Expected value to 5, is %d", t22.A)
	}

}

var testValuesQueue = []interface{}{
	&T2{A: 1, B: 2},
	&T2{A: 3, B: 4},
	&T2{A: 5, B: 6},
}

func TestCallbackParser(t *testing.T) {
	count := 0
	callback := func(vals interface{}) {
		//fmt.Println(vals)
		count++
	}

	config := TextTemplateParserConfig{
		Result: NewResultCallback(callback),
	}

	parser, err := NewParser(readFile("testrules/rules_callback.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewSimpleExecutor(parser)
	executor.Execute(testValuesQueue...)
	if count != 2 {
		log.Fatalf("Expected 2 callbacks, got %d", count)
	}

}

func TestQueueParser(t *testing.T) {

	in := make(chan interface{})
	out := make(chan interface{})

	config := TextTemplateParserConfig{
		Result: NewResultQueue(),
	}

	parser, err := NewParser(readFile("testrules/rules_queue.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewQueueExecutor(parser)
	executor.Execute(in, out)

	//writer
	go func(in chan interface{}, values []interface{}) {

		for _, v := range values {
			in <- v
		}

	}(in, testValuesQueue)

	expectedResults := 3

read:
	for {
		select {
		case v := <-out:
			expectedResults--
			switch v.(type) {
			case T2:
				// do something
				//fmt.Println(tv)
			}

			if expectedResults == 0 {
				close(in)
				close(out)
				break read
			}

			if expectedResults < 0 {
				log.Fatalf("received  %d more results", -1*expectedResults)
			}

		case <-time.After(time.Second * 3):
			log.Fatalf("received  %d less results", expectedResults)
		}
	}

}

var primitiveValues = []interface{}{
	[]string{"a,b,c"},
	[]int{1, 2},
	[]bool{true, false},
	[]float64{1.2, 1.3},
	1,
	true,
	1.3,
	"hello",
	map[string]string{
		"hello": "world",
	},
	map[string]bool{
		"hello": true,
	},
	map[string]interface{}{
		"err": errors.New(""),
	},
}

func TestPrimitives(t *testing.T) {

	for _, v := range primitiveValues {
		config := TextTemplateParserConfig{}

		parser, err := NewParser(readFile("testrules/rules_simple.xml"), config)
		if err != nil {
			log.Fatal(err)
		}

		executor := NewSimpleExecutor(parser)
		executor.Execute(v)

	}

}

func TestMapValue(t *testing.T) {
	m := map[string]interface{}{
		"T2": &T2{A: 1, B: 2},
	}

	config := TextTemplateParserConfig{}

	parser, err := NewParser(readFile("testrules/rules_map.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewSimpleExecutor(parser)
	executor.Execute(m)

}

func TestFilterTypes(t *testing.T) {
	t2 := &T2{A: 1, B: 2}

	config := TextTemplateParserConfig{}

	parser, err := NewParser(readFile("testrules/rules_filtertypes.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewSimpleExecutor(parser)
	executor.Execute(t2)

}

func TestQueueParserClose(t *testing.T) {

	in := make(chan interface{})
	out := make(chan interface{})

	config := TextTemplateParserConfig{
		Result: NewResultQueue(),
	}

	parser, err := NewParser(readFile("testrules/rules_queue.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := NewQueueExecutor(parser)
	executor.Execute(in, out)

	//writer
	go func(in chan interface{}, values []interface{}) {

		for _, v := range values {
			in <- v
		}

	}(in, testValuesQueue)

	time.Sleep(time.Millisecond * 100)

	close(in)
	executor.CloseResult()

}

func TestRulesetRequiredAttr(t *testing.T) {
	_, err := NewParser(readFile("testrules/rules_required_dataKey.xml"))
	if err == nil {
		log.Fatal(err)
	}

	_, err = NewParser(readFile("testrules/rules_required_filterTypes.xml"))
	if err == nil {
		log.Fatal(err)
	}
}

func TestUserFuncs(t *testing.T) {
	add := func(a, b int) int {
		return a + b
	}

	count := 0
	callback := func(vals interface{}) {
		//fmt.Println(vals)
		switch v := vals.(type) {
		case int:
			if v != 3 {
				log.Fatalf("add user func not called")
			}
		}

		count++
	}

	config := TextTemplateParserConfig{
		Result: NewResultCallback(callback),
		Userfuncs: template.FuncMap{
			"add": add,
		},
	}

	parser, err := NewParser(readFile("testrules/rules_userfuncs.xml"), config)
	if err != nil {
		log.Fatal(err)
	}
	t2 := &T2{A: 1, B: 2}
	executor := NewSimpleExecutor(parser)
	executor.Execute(t2)
	if count != 2 {
		log.Fatalf("Expected 2 callbacks, got %d", count)
	}

	config = TextTemplateParserConfig{
		Result: NewResultCallback(callback),
		Userfuncs: template.FuncMap{
			"_%f": add,
		},
	}

	parser, err = NewParser(readFile("testrules/rules_userfuncs.xml"), config)
	if err == nil {
		log.Fatal(err)
	}

}

func TestBadXML(t *testing.T) {
	_, err := NewParser(readFile("testrules/rules_badxml.xml"))
	if err == nil {
		log.Fatal(err)
	}

}

func TestBadFilterTypes(t *testing.T) {
	_, err := NewParser(readFile("testrules/rules_bad_filterTypes.xml"))
	if err == nil {
		log.Fatal(err)
	}

}

func TestBadPrioritiesCount(t *testing.T) {
	_, err := NewParser(readFile("testrules/rules_bad_prioritiescount.xml"))
	if err != nil {
		log.Fatal(err)
	}

}

func TestNoValues(t *testing.T) {
	parser, err := NewParser(readFile("testrules/rules_simple.xml"))
	if err != nil {
		log.Fatal(err)
	}

	executor := NewSimpleExecutor(parser)
	executor.Execute()

}

func BenchmarkSimpleParser(b *testing.B) {
	config := TextTemplateParserConfig{
		LogLevel: "fatal",
	}
	parser, err := NewParser(readFile("testrules/rules_simple.xml"), config)
	if err != nil {
		log.Fatal(err)
	}
	//t1 := &T1{T: &T{A: 1, B: 2}}
	t2 := &T2{A: 1, B: 2}
	executor := NewSimpleExecutor(parser)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		executor.Execute(t2)
	}
}
