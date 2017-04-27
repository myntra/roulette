package roulette

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"text/template"
)

// Parser holds the rules from a rule file
type Parser struct {
	Name           xml.Name `xml:"roulette"`
	Children       []*Rules `xml:"rules"`
	defaultFuncs   template.FuncMap
	userfuncs      template.FuncMap
	sortedChildren map[string]*Rules
	delimLeft      string
	delimRight     string
}

func (p *Parser) compile() {
	for _, rules := range p.Children {
		p.sortedChildren[rules.TypeName] = rules
	}

	for typeName, rules := range p.sortedChildren {

		for _, rule := range rules.Children {
			rule.compile(p.delimLeft, p.delimRight, p.defaultFuncs, p.userfuncs)
		}
		// sort by rule priority
		sort.Sort(p.sortedChildren[typeName])
	}
}

// Execute executes all rules in order of priority.
func (p *Parser) Execute(vals ...interface{}) error {
	typeName := p.buildMultiTypeName(vals)
	return p.multiTypesResult(typeName, false, vals)
}

// ExecuteOne executes in order of priority until a high priority rule is successful, after which rule
// execution stops.
func (p *Parser) ExecuteOne(vals ...interface{}) error {
	typeName := p.buildMultiTypeName(vals)
	return p.multiTypesResult(typeName, true, vals)
}

func getType(myvar interface{}) string {
	valueOf := reflect.ValueOf(myvar)

	if valueOf.Type().Kind() == reflect.Ptr {
		return reflect.Indirect(valueOf).Type().Name()
	}
	return valueOf.Type().Name()
}

func (p *Parser) buildMultiTypeName(vals interface{}) string {
	var typeName string
	for _, v := range vals.([]interface{}) {
		typeName = typeName + getType(v) + ","
	}
	typeName = strings.TrimRight(typeName, ",")
	return typeName
}

func (p *Parser) multiTypesResult(typeName string, one bool, vals interface{}) error {
	rules, ok := p.sortedChildren[typeName]
	if !ok {
		return fmt.Errorf("No rules found for types:%s", typeName)
	}

	if rules.DataKey == "" {
		return fmt.Errorf("No dataKey found for types:%s", typeName)
	}

	var tmplData map[string]interface{}

	valsData := make(map[string]interface{})
	for _, val := range vals.([]interface{}) {
		valsData[getType(val)] = val
	}

	tmplData = valsData

	if len(vals.([]interface{})) > 1 {
		tmplData[rules.DataKey] = valsData
	}

	for _, rule := range rules.Children {
		// execute rules
		if rule.Template == nil {
			log.Println(errors.New("rule expression not found"))
			continue
		}

		var buf bytes.Buffer
		err := rule.Template.Execute(&buf, tmplData)
		if err != nil {
			log.Println(err)
			continue
		}

		var result bool
		err = json.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			log.Println(err)
			continue
		}

		// first high priority rule successful, break
		if result && one {
			break
		}

	}

	return nil
}

// AddFuncs allows additional functions to be added to the parser
// Functions must be of the signature: f(arg1,arg2, prevVal ...string)string
// See funcmap.go for examples.
func (p *Parser) AddFuncs(funcMap template.FuncMap) error {
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
		p.userfuncs[name] = fn
	}
	p.compile()

	return nil
}

// RemoveFuncs removes previously added functions from the parser
func (p *Parser) RemoveFuncs(funcMap template.FuncMap) {
	for k := range funcMap {
		delete(p.userfuncs, k)
	}

	p.compile()
}

// Delims sets the custom delimieters for parsing the text/template expression
func (p *Parser) Delims(left, right string) {
	p.delimLeft = left
	p.delimRight = right
	p.compile()
}

// Update is a wrapper over new for the current parser
// The method recompiles the templates.
func (p *Parser) Update(data []byte) error {
	p, err := New(data)
	if err != nil {
		return err
	}
	return nil
}

// New returns a new roulette format xml parser
func New(data []byte) (*Parser, error) {

	parser := &Parser{
		delimLeft:      delimLeft,
		delimRight:     delimRight,
		sortedChildren: make(map[string]*Rules),
		defaultFuncs:   defaultFuncMap,
		userfuncs:      template.FuncMap{},
	}

	err := xml.Unmarshal(data, parser)
	if err != nil {
		return nil, err
	}

	parser.compile()

	return parser, nil
}
