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

- Builtin functions for writing simple rule expressions. 
- Supports injecting custom functions.
- Can namespace a set of rules for custom `types`.
- Allows setting priority of a `rule`.


This pacakge is used for firing business actions based on a textual decision tree. It uses the powerful control structures in `text/template` and xml parsing from `encoding/xml` to build the tree from a `roulette` xml file.

### go get
```
$ go get github.com/myntra/roulette
```

### Usage:

From `testrules/rules.xml`

```xml
<roulette>
    <rules types="Person,Company" dataKey="MyData">        
        <rule name="setAge" priority="2">
                <r>with .MyData</r>
                    <r>
                       ge .Person.Experience 7 |
                       in .Person.Age 15 30 |
                       le .Person.Vacations 5 | 
                       eq .Person.Position "SSE" |
                       eq .Company.Name "Myntra" | 
                       .Person.SetAge 25 
                    </r>
                <r>end</r>
        </rule>

        <rule name="setSalary" priority="1">
                <r>with .MyData</r>
                    <r>with .Person </r>
                        <r>
                             eq .Position "SSE" |
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

#### XML Tags

- `roulette` : the root tag.

- `rules`: a types namespaced tag with `rule` children. The attributes `types` and `dataKey` are **required**. `types` value can be a single type or a comma separated list of types. To match `rules` set, atleast one of the types from this list should be given to `parse.Execute` or `parse.ExecuteOne`.

- `rule`: tag which holds the `rule expression`. The attributes `name` and `priority` are **optional**. The default value of `priority` is 0. There is no guarantee for order of execution if `priority` is not set.


```xml
<roulette>
    <rules types="Person,Company" dataKey="MyData">        
        <rule name="setAge" priority="2">
                <r>with .MyData</r>
                    <r>
                       le .Person.Vacations 5 |
                       and (gt .Person.Experience 6) (in .Person.Age 15 30) |
                       or (eq .Person.Position "SSE") (eq .Company.Name "Myntra") |
                       .Person.SetAge 25
                    </r>
                <r>end</r>
        </rule>

        <rule name="setSalary" priority="1">
                <r>with .MyData</r>
                    <r>with .Person </r>
                        <r>
                             eq .Position "SSE" | .SetSalary 50000
                        </r>
                    <r>end</r>
                <r>end</r>
        </rule>
    </rules>
</roulette>
```

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
- The pipe `|` operator takes a previously evaluated value and passes it to the next function as the last argument.
- For more information on go templating: [text/template](https://golang.org/pkg/text/template/)



#### Builtin Functions

Apart from the built-in functions from `text/template`, the following functions are available.

Default functions reside in `funcmap.go`. They have been sourced and modified from `src/text/template/funcs.go` to make them work with pipelines and keep the expression uncluttered.
The idea is to keep the templating language readable and easy to write.

| Function      | Usage         |
| ------------- |:-------------:|
| in           |  `val >= minVal && val <= maxval`, e.g. `in 2 1 4` => `true` | 
| gt           |  `> op`, e.g. `gt 1 2` |
| ge           |  `>= op`, e.g. `ge 1 2` |
| lt           |  `<= op`, e.g. `lt 1 2` |
| le           |  `< op`, e.g. `le 1 2` |
| eq           |  `== op`, e.g. `eq "hello" "world"` |
| ne           |  `!= op`, e.g.`ne "hello" "world"` |
| not		   | `!op` , e.g.`not 1`|
| and 		   | `op1 && op2`, e.g. `and (expr1) (expr2)`|
| or 		   | `op1 // op2`, e.g. `or (expr1) (expr2)`|
| `|` pipe 		   | `the output of fn1 is the last argument of fn2`, e.g. `fn1 | fn2`|


#### TODO
- More builtin functions.
- More tests.
- More examples: roulette templates and go code.
- Static checker(?).

#### Attribution
The `roulette.png` image is sourced from https://thenounproject.com/term/roulette/143243/ with a CC license.