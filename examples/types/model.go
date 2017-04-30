package types

import (
	"io/ioutil"
	"log"
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
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}
	p.Age = age
	return true
}

// SetSalary ...
func (p *Person) SetSalary(salary int, prevVal ...bool) bool {
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
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

func readFile(path string) []byte {
	ruleFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ruleFile read err #%v  at path %v", err, path)
	}

	return ruleFile
}
