package roulette

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"text/template"
)

// default delimeters
var delimLeft = "<r>"
var delimRight = "</r>"

// Parser interface provides methods for the executor to update the tree, execute the values and get result.
type Parser interface {
	Update(data []byte) error
	Execute(vals interface{})
	Result() Result
}

// TextTemplateParser holds the rules from a rule file
type TextTemplateParser struct {
	Name     xml.Name               `xml:"roulette"`
	Rulesets []*TextTemplateRuleset `xml:"ruleset"`

	DefaultFuncs    template.FuncMap
	Userfuncs       template.FuncMap
	DelimLeft       string
	DelimRight      string
	WorkflowPattern string // filter rulesets based on the pattern
	RuleResult      Result
	Get             chan interface{}
}

// Compile compiles the parser's rulesets
func (p *TextTemplateParser) Compile() error {

	var _ Ruleset = &TextTemplateRuleset{}
	for _, ruleset := range p.Rulesets {

		ruleset.Result(p.RuleResult)

		err := ruleset.Compile(p.DelimLeft, p.DelimRight, p.WorkflowPattern, p.DefaultFuncs, p.Userfuncs)
		if err != nil {
			return err
		}

		ruleset.Sort()
	}

	return nil
}

// Execute executes the parser's rulesets
func (p *TextTemplateParser) Execute(vals interface{}) {
	for _, ruleset := range p.Rulesets {
		ruleset.Execute(vals)
	}
}

// Result returns the parser's result.
func (p *TextTemplateParser) Result() Result {
	return p.RuleResult
}

// AddFuncs allows additional functions to be added to the parser
// Functions must be of the signature: f(arg1,arg2, prevVal ...string)string
// See funcmap.go for examples.
func (p *TextTemplateParser) AddFuncs(funcMap template.FuncMap) error {
	for name, fn := range funcMap {
		if !goodName(name) {
			return fmt.Errorf("function name %s is not a valid identifier", name)
		}
		v := reflect.ValueOf(fn)
		if v.Kind() != reflect.Func {
			return fmt.Errorf("value for " + name + " not a function")
		}
		if !goodFunc(v.Type()) {
			return fmt.Errorf("can't install method/function %q with %d results", name, v.Type().NumOut())
		}
		p.Userfuncs[name] = fn
	}

	return p.Compile()
}

// RemoveFuncs removes previously added functions from the parser
func (p *TextTemplateParser) RemoveFuncs(funcMap template.FuncMap) {
	for k := range funcMap {
		delete(p.Userfuncs, k)
	}

	p.Compile()
}

// Delims sets the custom delimieters for parsing the text/template expression
func (p *TextTemplateParser) Delims(left, right string) {
	p.DelimLeft = left
	p.DelimRight = right
	p.Compile()
}

// Update is a wrapper over new for the current parser
// The method recompiles the templates.
func (p *TextTemplateParser) Update(data []byte) error {

	err := xml.Unmarshal(data, p)
	if err != nil {
		return err
	}

	// validate
	err = p.Compile()
	if err != nil {
		return err
	}

	return nil
}

// NewTextTemplateParser returns a new roulette format xml parser.
func NewTextTemplateParser(data []byte, result Result, workflowPattern string) (Parser, error) {

	get := make(chan interface{})

	parser := &TextTemplateParser{
		DelimLeft:       delimLeft,
		DelimRight:      delimRight,
		DefaultFuncs:    defaultFuncMap,
		Userfuncs:       template.FuncMap{},
		Get:             get,
		WorkflowPattern: workflowPattern,
		RuleResult:      result,
	}

	err := parser.Update(data)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

// NewSimpleParser returns a TextTemplateParser with a nil Result.
func NewSimpleParser(data []byte, pattern string) (Parser, error) {
	return NewTextTemplateParser(data, nil, pattern)
}

// NewCallbackParser returns a TextTemplateParser with a new ResultCallback
func NewCallbackParser(data []byte, fn func(interface{}), pattern string) (Parser, error) {
	return NewTextTemplateParser(data, NewResultCallback(fn), pattern)
}

// NewQueueParser returns a TextTemplateParser with a new ResultQueue.
func NewQueueParser(data []byte, pattern string) (Parser, error) {
	return NewTextTemplateParser(data, NewResultQueue(), pattern)
}
