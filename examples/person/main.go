package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/myntra/roulette"
)

// Person ...
type Person struct {
	ID         int
	Age        int
	Experience int
	Vacations  int
	Position   string
}

// SetAge ...
func (p *Person) SetAge(age ...string) bool {
	p.Age = 25
	return true
}

// Company ...
type Company struct {
	Name string
}

func getParser(path string) *roulette.Parser {
	ruleFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ruleFile read err #%v  at path %v", err, "test_rule.xml")
	}

	parser, err := roulette.New(ruleFile)
	if err != nil {
		panic(err)
	}

	return parser
}

func main() {

	p := Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}
	c := Company{Name: "Myntra"}

	parser := getParser("../../testrules/test_rule.xml")

	// add custom functions
	parser.AddFuncs(template.FuncMap{
		"customFunc": customFunc,
	})

	// get only the top priority result
	ruleResult, err := parser.ResultOne(p)
	if err != nil {
		log.Fatal(err)
	}
	if ruleResult.Name() != "setAgeField" {
		log.Fatal("top priority rule was not returned")
	}

	v, ok := ruleResult.Val().(*Person)
	if !ok {
		log.Fatal("Incorrect type returned")
	}

	if v.Age != 25 {
		log.Fatal("Age field was not set")
	}

	// get all results result
	ruleResults := parser.ResultAll(p)
	if err != nil {
		log.Fatal(err)
	}

	for _, ruleResult := range ruleResults {
		if ruleResult.Name() == "setAgeField" {
			fmt.Println(ruleResult.Name(), ruleResult.Val().(*Person))
			continue
		}
		fmt.Println(ruleResult.Name(), ruleResult.BoolVal())
	}

	// use company type
	ruleResult, err = parser.ResultOne(c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ruleResult.Name(), ruleResult.BoolVal())

	// modify person
	p2 := Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}
	parser = getParser("../../testrules/test_rule_type_method.xml")
	parser.Execute(&p2)
	fmt.Println("updated", p2)

}

// this function signature is required:
// f(arg1,arg2, prevVal ...string)string
func customFunc(val1 interface{}, val2 string, prevVal ...string) string {
	fmt.Println("customFunc trigerred", val1, val2, prevVal)
	return "true"
}
