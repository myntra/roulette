package roulette

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

type Person struct {
	ID         int
	Age        int
	Experience int
	Vacations  int
	Position   string
}

// Company ...
type Company struct {
	Name string
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var p = Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}

func getParser(path string) *Parser {
	ruleFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ruleFile read err #%v  at path %v", err, "test_rule.xml")
	}

	parser, err := New(ruleFile)
	check(err)

	return parser
}

func TestRuleFile(t *testing.T) {

	//c := Company{Name: "myntra"}
	parser := getParser("testrules/test_rule.xml")

	ruleResults := parser.ResultAll(p)

	for _, result := range ruleResults {

		fmt.Println(result.Name())
	}

}

func TestRuleSetField(t *testing.T) {
	parser := getParser("testrules/test_rule_setfield.xml")
	// get top priority result
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

}
