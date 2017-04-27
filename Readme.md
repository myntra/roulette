<p align="center">
  <img src="https://cdn.rawgit.com/myntra/roulette/master/images/roulette.png" height="118" width="130" />

  <h3 align="center">Roulette</h3>
  <p align="center">A text/template based package which triggers actions from rules defined in an xml file.</p>
  <p align="center">
  	<img src="http://badges.github.io/stability-badges/dist/experimental.svg"/>
    <a href="https://travis-ci.org/myntra/roulette"><img src="https://travis-ci.org/myntra/roulette.svg?branch=master"></a>
    <a href="https://godoc.org/github.com/myntra/roulette"><img src="https://godoc.org/github.com/myntra/roulette?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/myntra/roulette"><img src="https://goreportcard.com/badge/github.com/myntra/roulette"></a>
  </p>
</p>

---
### Features:

- Powerful `text/template` control structures.
- Builtin functions. 
- Supports injecting custom functions.
- Can namespace a set of rules for custom `types`.


This pacakge is used for firing business actions based on a textual decision tree. It uses the powerful control structures in `text/template` and `encoding/xml` to build the tree from an rules xml file.

### go get
```
$ go get github.com/myntra/roulette
```

### Usage:

#### Defining Rules in XML:

- Write valid `text/template` control structures within the `<rule>...</rule>` tag.
- Namespace rules by custom types. e.g: 

	`<rules types="Person,Company">...</rules>`

- Set `priority` of rules within namespace `types`.
- Add custom functions to the parser using the method `parser.AddFuncs`. The function must have the signature:
	
	`func(arg1,...,argN,prevVal ...bool)bool`
 
  to allow rule execution status propagation.
- Methods to be invoked from the rules file must also be of the above signature.
- Invalid/Malformed rules are skipped and the error is logged.
- For more information on go templating: [text/template](https://golang.org/pkg/text/template/)

From `testrules/rules.xml`

```xml
<roulette>
    <rules types="Person,Company" dataKey="MyData">        
        <rule name="setAge" priority="2">
                <r>with .MyData</r>
                    <r>
                       gte .Person.Experience 7 |
                       within .Person.Age 15 30 |
                       lte .Person.Vacations 5 | 
                       eql .Person.Position "SSE" |
                       eql .Company.Name "Myntra" | 
                       .Person.SetAge 25 
                    </r>
                <r>end</r>
        </rule>

        <rule name="setSalary" priority="1">
                <r>with .MyData</r>
                    <r>with .Person </r>
                        <r>
                             eql .Position "SSE" |
                            .SetSalary 50000
                        </r>
                    <r>end</r>
                <r>end</r>
        </rule>
    </rules>
</roulette>
```

#### API:

`parser.Execute` : Applies all matching rules for the types namespace in order of `priority`. 


From `examples/person/main.go`

```go
...

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

func main() {
	p := Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}
	c := Company{Name: "Myntra"}

	// execute all rules
	parser := getParser("testrules/rules.xml")
	err := parser.Execute(&p, &c)
	if err != nil {
		log.Fatal(err)
	}
}

....

```

`parser.ExecuteOne`: Applies matching rules for the types namespace in order of `priority` until a rule is successful. 


```go

p := Person{ID: 1, Age: 20, Experience: 7, Vacations: 4, Position: "SSE"}
	c := Company{Name: "Myntra"}

	// execute rules until successful
	parser := getParser("testrules/rules.xml")
	err := parser.ExecuteOne(&p, &c)
	if err != nil {
		log.Fatal(err)
	}

```



#### Builtin Functions

Apart from the built-in functions from `text/template`, the following functions are available. Some functions have been overidden to make their output parseable.

Default functions reside in `funcmap.go`

| Function      | Usage         | Signature  |
| ------------- |:-------------:| -----:|
| within           |  val >= minVal && val <= maxval, `within 2 1 4` | `within(fieldVal int, minVal int, maxVal int, prevVal ...bool) bool`
| gte           |  >= op, `gte 1 2` | `gte(fieldVal int, minVal int, prevVal ...bool) bool`
| lte           |  <= op, `lte 1 2` | `lte(fieldVal int, maxVal int, prevVal ...bool) bool` |
| eql           |  == op, `eq "hello" "world"` |  `eql(fieldVal interface{}, val interface{}, prevVal ...bool) bool` |


#### TODO
- More builtin functions.
- More tests.
- More examples: roulette templates and go code.
- Static checker.

#### Attribution
The `roulette.png` image is sourced from https://thenounproject.com/term/roulette/143243/ with a CC license.