package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

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

var testValuesQueue = []interface{}{
	types.Person{ID: 1, Age: 32, Experience: 7, Vacations: 4, Position: "SSE"}, //fail
	types.Person{ID: 2, Age: 20, Experience: 7, Vacations: 6, Position: "SSE"}, //fail
	types.Person{ID: 3, Age: 20, Experience: 7, Vacations: 5, Position: "SSE"}, //pass
	[]interface{}{
		types.Person{ID: 4, Age: 20, Experience: 7, Vacations: 5, Position: "SSE"}, // pass
		types.Company{Name: "Myntra"},                                              // pass
	},
}

func main() {
	//    le .Person.Vacations 5 |
	//    and (gt .Person.Experience 6) (in .Person.Age 15 30) |
	//    eq .Person.Position "SSE"  |
	//    .result.Put .Person

	//	done := make(chan struct{})
	in := make(chan interface{})
	out := make(chan interface{})

	// get rule results on a queue
	parser, err := roulette.NewQueueParser(readFile("../rules.xml"), "")
	if err != nil {
		log.Fatal(err)
	}

	executor := roulette.NewQueueExecutor(parser)
	executor.Execute(in, out)

	//writer
	go func(in chan interface{}, values []interface{}) {

		for _, v := range values {
			in <- v
		}

	}(in, testValuesQueue)

	expectedResults := 2

read:
	for {
		select {
		case v := <-out:
			expectedResults--
			fmt.Println(v)
			switch tv := v.(type) {
			case types.Person:
				// do something
				if !(tv.ID == 4 || tv.ID == 3) {
					log.Fatal("Unexpected Result", tv)
				}
			}

			if expectedResults == 0 {
				break read
			}

			if expectedResults < 0 {
				log.Fatalf("received  %d more results", -1*expectedResults)
			}

		case <-time.After(time.Second * 5):
			log.Fatalf("received  %d less results", expectedResults)
		}
	}

}
