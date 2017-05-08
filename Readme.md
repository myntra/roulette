<p align="center">
  <img src="https://cdn.rawgit.com/myntra/roulette/master/images/roulette.png" height="118" width="130" />

  <h3 align="center">Roulette</h3>
  <p align="center">A text/template based package which triggers actions from rules defined in an xml file.</p>
  <p align="center">
  	<img src="http://badges.github.io/stability-badges/dist/experimental.svg"/>
    <a href='https://coveralls.io/github/myntra/roulette?branch=master'><img src='https://coveralls.io/repos/github/myntra/roulette/badge.svg?branch=master' alt='Coverage Status' /></a>
    <a href="https://travis-ci.org/myntra/roulette"><img src="https://travis-ci.org/myntra/roulette.svg?branch=master"></a>
    <a href="https://godoc.org/github.com/myntra/roulette"><img src="https://godoc.org/github.com/myntra/roulette?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/myntra/roulette"><img src="https://goreportcard.com/badge/github.com/myntra/roulette"></a>
  </p>
</p>

---
<!-- TOC -->

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
    - [Overview](#overview)
- [Guide](#guide)
    - [Roulette XML file:](#roulette-xml-file)
        - [Tags and Attributes](#tags-and-attributes)
            - [Roulette](#roulette)
            - [Ruleset](#ruleset)
            - [Rule](#rule)
            - [Rule Expressions](#rule-expressions)
        - [Defining Rules in XML](#defining-rules-in-xml)
    - [Parsers](#parsers)
        - [TextTemplateParser](#texttemplateparser)
    - [Results](#results)
        - [ResultCallback](#resultcallback)
        - [ResultQueue](#resultqueue)
    - [Executors](#executors)
        - [SimpleExecutor](#simpleexecutor)
            - [SimpleExecutor with Callback](#simpleexecutor-with-callback)
        - [QueueExecutor](#queueexecutor)
- [Builtin Functions](#builtin-functions)
- [Attributions](#attributions)

<!-- /TOC -->

## Features

- Builtin functions for writing simple rule expressions. 
- Supports injecting custom functions.
- Can namespace a set of rules for custom `types`.
- Allows setting priority of a `rule`.


This pacakge is used for firing business actions based on a textual decision tree. It uses the powerful control structures in `text/template` and xml parsing from `encoding/xml` to build the tree from a `roulette` xml file.

## Installation
```
$ go get github.com/myntra/roulette
```

## Usage
### Overview

From `examples/rules.xml`

```xml
<roulette>
    <!--filterTypes="T1,T2,T3..."(required) allow one or all of the types for the rules group. * pointer filterting is not done .-->
    <!--filterStrict=true or false. rules group executed only when all types are present -->
    <!--prioritiesCount= "1" or "2" or "3"..."all". if 1 then execution stops after "n" top priority rules are executed. "all" executes all the rules.-->
    <!--dataKey="string" (required) root key from which user data can be accessed. -->
    <!--resultKey="string" key from which result.put function can be accessed. default value is "result".-->
    <!--workflow: "string" to group rulesets to the same workflow.-->

    <ruleset name="personRules" dataKey="MyData" resultKey="result" filterTypes="types.Person,types.Company" 
        filterStrict="false" prioritiesCount="all"  workflow="promotioncycle">

        <rule name="personFilter1" priority="3">
                <r>with .MyData</r>
                    <r>
                       le .types.Person.Vacations 5 |
                       and (gt .types.Person.Experience 6) (in .types.Person.Age 15 30) |
                       eq .types.Person.Position "SSE" |
                       .types.Person.SetAge 25
                    </r>
                <r>end</r>
        </rule>

        <rule name="personFilter2" priority="2">
                <r>with .MyData</r>
                    <r>
                       le .types.Person.Vacations 5 |
                       and (gt .types.Person.Experience 6) (in .types.Person.Age 15 30) |
                       eq .types.Person.Position "SSE"  |
                       .result.Put .types.Person
                    </r>
                <r>end</r>
        </rule>

        <rule name="personFilter3" priority="1">
            <r>with .MyData</r>
                    <r>
                    le .types.Person.Vacations 5 |
                    and (gt .types.Person.Experience 6) (in .types.Person.Age 15 30) |
                    eq .types.Person.Position "SSE"  |
                    eq .types.Company.Name "Myntra" |
                     .result.Put .types.Company |
                    </r>
            <r>end</r>
        </rule>
    </ruleset>

    <ruleset name="personRules2" dataKey="MyData" resultKey="result" filterTypes="types.Person,types.Company" 
    filterStrict="false" prioritiesCount="all" workflow="demotioncycle">
    <rule name="personFilter1" priority="1">
        <r>with .MyData</r>
            <r>
                eq .types.Company.Name "Myntra" | .types.Person.SetSalary 30000
            </r>
        <r>end</r>
    </rule>
    </ruleset>    
</roulette>
```

From `examples/...`

`simple`

```go
...
   	p := types.Person{ID: 1, Age: 20, Experience: 7, Vacations: 5, Position: "SSE"}
	c := types.Company{Name: "Myntra"}

	config := roulette.TextTemplateParserConfig{}

	parser, err := roulette.NewParser(readFile("../rules.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := roulette.NewSimpleExecutor(parser)
	executor.Execute(&p, &c, []string{"hello"}, false, 4, 1.23)

	if p.Age != 25 {
		log.Fatal("Expected Age to be 25")
	}


  ...
```

`workflows`

```go
...

	p := types.Person{ID: 1, Age: 20, Experience: 7, Vacations: 5, Position: "SSE"}
	c := types.Company{Name: "Myntra"}

	config := roulette.TextTemplateParserConfig{
		WorkflowPattern: "demotion*",
	}
	// set the workflow pattern
	parser, err := roulette.NewParser(readFile("../rules.xml"), config)
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
...
```


`callback`

```go
...
    count := 0
	callback := func(vals interface{}) {
		fmt.Println(vals)
		count++
	}

	config := roulette.TextTemplateParserConfig{
		Result: roulette.NewResultCallback(callback),
	}

	parser, err := roulette.NewParser(readFile("../rules.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := roulette.NewSimpleExecutor(parser)
	executor.Execute(testValuesCallback...)
	if count != 2 {
		log.Fatalf("Expected 2 callbacks, got %d", count)
	}
...
```
`queue`

```go
...
in := make(chan interface{})
	out := make(chan interface{})

	config := roulette.TextTemplateParserConfig{
		Result: roulette.NewResultQueue(),
	}

	// get rule results on a queue
	parser, err := roulette.NewParser(readFile("../rules.xml"), config)
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
...
```

## Guide
### Roulette XML file: 

#### Tags and Attributes

##### Roulette
  
  `roulette` is the root tag of the xml. It could contain a list of `ruleset` tags. 

##### Ruleset

`ruleset`: a types namespaced tag with `rule` children. The attributes `filterTypes` and `dataKey` are **required**. To match `ruleset` , atleast one of the types from this list should be an input for the executor. 

`Attributes`: 

- `filterTypes`: "T1,T2,T3..."(required) allow one or all of the types for the rules group. * pointer filterting is not done.

- `filterStrict`: true or false. rules group executed only when all types are present.

- `prioritiesCount`: "1" or "2" or "3"..."all". if 1 then execution stops after "n" top priority rules are executed. "all" executes all the rules

- `dataKey`: "string" (required) root key from which user data can be accessed.

- `resultKey`: "string" key from which result.put function can be accessed. default value is "result".
- `workflow`: "string" to group rulesets to the same workflow. The parser can then be created with a wildcard pattern to filter out rilesets.  "*", "?" glob pattern matching is expected.


##### Rule

The tag which holds the `rule expression`. The attributes `name` and `priority` are **optional**. The default value of `priority` is 0. There is no guarantee for order of execution if `priority` is not set.

`Attributes`:

- `name`: name of the rule.

- `priority`: priority rank of the rule within the ruleset. 


##### Rule Expressions

Valid `text/template` expression. The delimeters can be changed from the default `<r></r>` using the parse api.

#### Defining Rules in XML

- Write valid `text/template` control structures within the `<rule>...</rule>` tag.
- Namespace rules by custom types. e.g: 

	`<ruleset filterTypes="Person,Company">...</ruleset>`

- Set `priority` of rules within namespace `filterTypes`.
- Add custom functions to the parser using the method `parser.AddFuncs`. The function must have the signature:
	
	`func(arg1,...,argN,prevVal ...bool)bool`
 
  to allow rule execution status propagation.
- Methods to be invoked from the rules file must also be of the above signature.
- Invalid/Malformed rules are skipped and the error is logged.
- The pipe `|` operator takes a previously evaluated value and passes it to the next function as the last argument.
- For more information on go templating: [text/template](https://golang.org/pkg/text/template/)


### Parsers

#### TextTemplateParser
Right now the package provides a single parser: `TextTemplateParser`. As the name suggests the parser is able to read xml wrapped over a valid `text/template` expression and executes it.


### Results

Types which implements the `roulette.Result`. If a parser is initalised with a `Result` type, rule expressions with `result.Put` become valid. `result.Put` function accepts an `interface{}` type as a parameter.

#### ResultCallback

`roulette.ResultCallback`: An implementation of the `roulette.Result` interface which callbacks the provided function with `result.Put` value.

#### ResultQueue 

`roulette.ResultQueue`: An implementation of the `roulette.Result` interface which puts the value received from `result.Put` on a channel.

### Executors

An executor takes a parser and tests an incoming values against the rulesets. Executors implement the `roulette.SimpleExecute` and `roulette.QueueExecute` interfaces. The result is then caught by a struct which implements the `roulette.Result` interface. 

#### SimpleExecutor

A simple implementation of the `roulette.SimpleExecute` interface which has a `parser` with  `nil` `Result` set. This is mainly used to directly modify the input object. The executor ignores rule expressions which contain `result.Put`.

```go

parser,err := NewTextTemplateParser(data, nil,"")
// or parser, err := roulette.NewSimpleParser(data,nil,"")
executor := roulette.NewSimpleExecutor(parser)
executor.Execute(t1,t2)
```

##### SimpleExecutor with Callback

An implementation of the `roulette.SimpleExecute` interface. which accepts a parser initialized with `roulette.ResultCallback`.

```go
    config := roulette.TextTemplateParserConfig{}

	parser, err := roulette.NewParser(readFile("../rules.xml"), config)
	if err != nil {
		log.Fatal(err)
	}

	executor := roulette.NewSimpleExecutor(parser)
	executor.Execute(...)
```

#### QueueExecutor 

An implementation of the `roulette.QueueExecute` interface. which accepts the `roulette.ResultQueue` type. The `Execute` method expects an input and an output channel to write values and read results respectively. 

```go

in := make(chan interface{})
		out := make(chan interface{})

		config := roulette.TextTemplateParserConfig{
			Result: roulette.NewResultQueue(),
		}

		// get rule results on a queue
		parser, err := roulette.NewParser(readFile("../rules.xml"), config)
		if err != nil {
			log.Fatal(err)
		}

		executor := roulette.NewQueueExecutor(parser)
		executor.Execute(in, out)
```

For concrete examples of the above please see the `examples` directory. 


## Builtin Functions

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
| result.Put   | `result.Put Value` where `result` is the defined `resultKey`|
 


 - pipe operator | : `Usage: the output of fn1 is the last argument of fn2`, e.g. `fn1 1 2| fn2 1 2 `

The functions from the excellent package [sprig](http://masterminds.github.io/sprig/) are also available.

## Attributions
The `roulette.png` image is sourced from https://thenounproject.com/term/roulette/143243/ with a CC license.