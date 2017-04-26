package main

import (
	"fmt"
	"io/ioutil"
	"log"

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

	ruleResult, err = parser.ResultOne(c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ruleResult.Name(), ruleResult.BoolVal())

}
