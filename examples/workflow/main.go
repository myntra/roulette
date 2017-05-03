package main

import (
	"io/ioutil"
	"log"

	"github.com/myntra/roulette"
	"github.com/myntra/roulette/examples/types"
)

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

func main() {

	p := types.Person{ID: 1, Age: 20, Experience: 7, Vacations: 5, Position: "SSE"}
	c := types.Company{Name: "Myntra"}

	// set the workflow pattern
	parser, err := roulette.NewSimpleParser(readFile("../rules.xml"), "demotion*")
	if err != nil {
		log.Fatal(err)
	}

	executor := roulette.NewSimpleExecutor(parser)
	executor.Execute(&p, &c, []string{"hello"}, false, 4, 1.23)

	if p.Salary != 30000 {
		log.Fatal("Expected Salary to be 30000")
	}

	if p.Age != 20 {
		log.Fatal("Expected Age to be 20")
	}

}
