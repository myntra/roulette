package roulette

import (
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
	Salary     int
}

func (p *Person) SetAge(age int, prevVal ...bool) bool {
	if !checkPrevVal(prevVal) {
		return false
	}
	p.Age = age
	return true
}

// SetSalary ...
func (p *Person) SetSalary(salary int, prevVal ...bool) bool {
	if !checkPrevVal(prevVal) {
		return false
	}
	p.Salary = salary
	return true
}

// Company ...
type Company struct {
	Name string
}

func checkPrevVal(prevVal []bool) bool {
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}
	return true
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getParser(path string) *Parser {
	ruleFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ruleFile read err #%v  at path %v", err, "test_rule.xml")
	}

	parser, err := New(ruleFile)
	check(err)

	return parser
}

func TestExecuteAllMultiType(t *testing.T) {
	p := Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}
	c := Company{Name: "Myntra"}

	// execute all rules
	parser := getParser("testrules/rules.xml")
	err := parser.Execute(&p, &c)
	if err != nil {
		log.Fatal(err)
	}

	if p.Age != 25 {
		log.Fatal("Expected Age to be set to 25")
	}
}

func TestExecuteOne(t *testing.T) {
	p := Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}
	c := Company{Name: "Myntra"}

	// execute all rules
	parser := getParser("testrules/rules.xml")
	err := parser.ExecuteOne(&p, &c)
	if err != nil {
		log.Fatal(err)
	}

	if p.Age == 25 {
		log.Fatal("Expected Age to be 20")
	}

	if p.Salary != 50000 {
		log.Fatal("Expected Salary to be 50000")
	}
}
