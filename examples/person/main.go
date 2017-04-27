package person

import (
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
	Salary     int
	Position   string
}

// SetAge ...
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkPrevVal(prevVal []bool) bool {
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}
	return true
}

func getParser(path string) *roulette.Parser {
	ruleFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ruleFile read err #%v  at path %v", err, "test_rule.xml")
	}

	parser, err := roulette.New(ruleFile)
	check(err)

	return parser
}

func main() {
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

	executeOne()
}

func executeOne() {

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
