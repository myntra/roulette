package main

import (
	"fmt"
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

var testValuesCallback = []interface{}{
	types.Person{ID: 1, Age: 20, Experience: 7, Vacations: 5, Position: "SSE"}, //pass
	types.Company{Name: "Myntra"},
}

func main() {

	count := 0
	callback := func(vals interface{}) {
		fmt.Println(vals)
		count++
	}

	parser, err := roulette.NewCallbackParser(readFile("../rules.xml"), callback, "")
	if err != nil {
		log.Fatal(err)
	}

	executor := roulette.NewSimpleExecutor(parser)
	executor.Execute(testValuesCallback...)
	if count != 2 {
		log.Fatalf("Expected 2 callbacks, got %d", count)
	}

}
