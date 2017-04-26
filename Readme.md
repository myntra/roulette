<p align="center">
  <h3 align="center">Roulette</h3>
  <p align="center">A package which returns values/takes actions based on the rules defined in an xml file. These values can be further used to trigger actions.</p>
  <p align="center">
    <a href="https://travis-ci.org/myntra/roulette"><img src="https://travis-ci.org/myntra/roulette.svg?branch=master"></a>
    <a href="https://godoc.org/github.com/myntra/roulette"><img src="https://godoc.org/github.com/myntra/roulette?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/myntra/roulette"><img src="https://goreportcard.com/badge/github.com/myntra/roulette"></a>
  </p>
</p>

---

This pacakge is based on `text/template`. It uses the powerful control structures in `text/template` to return actionable data. With some reflect magic, it's also able to return updated concrete types as shown in the example below.

### go get
```
$ go get github.com/myntra/roulette
```

### Usage

From `examples/person`

#### Define Rules in XML

```xml
<roulette>
    <rules type="Person">
        <rule name="ageWithinRange" resultType="bool" priority="4">
                <r>with .Person</r>
                    <r>within .Age 15 30</r>
                <r>end</r>               
        </rule>

        <rule  name="expWithinRange"  resultType="bool" priority="2">
                <r>with .Person</r>
                    <r>within .Experience 5 10</r>
                <r>end</r>
        </rule>

        <rule name="promote" resultType="bool" priority="3">
                <r>with .Person</r>
                    <r>gte .Experience 7 | within .Age 15 30 | lte .Vacations 5 | eql .Position "SSE"</r>
                <r>end</r>
        </rule>
        
        <rule name="setAgeField" resultType="Person" priority="1">
                <r>with .Person</r>
                    <r>gte .Experience 7 | within .Age 15 30 | lte .Vacations 5 | eql .Position "SSE" | set . "Age" 25 </r>
                <r>end</r>
        </rule>
    </rules>

    <rules type="Company">
        <rule name="ageWithinRange" resultType="bool">
                <r>with .Company</r>
                <r>eql .Name "Myntra" </r>
                <r>end</r>
        </rule>
    </rules>
</roulette>
```

#### Go


```go
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
```
