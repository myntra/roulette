package roulette

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"
	"reflect"
	"sort"
	"text/template"

	"github.com/fatih/structs"
)

// default delimeters
var delimLeft = "<r>"
var delimRight = "</r>"

// Rule is a single rule expression. A rule expression is a valid go text/template
type Rule struct {
	Name       string `xml:"name,attr"`
	ResultType string `xml:"resultType,attr"`
	Priority   int    `xml:"priority,attr"`
	Expr       string `xml:",innerxml"`
	Template   *template.Template
}

// compile initialises rule templates
func (r *Rule) compile(left, right string) {
	r.Template = template.Must(template.New(r.Name).Delims(left, right).Funcs(funcMap).Parse(r.Expr))
}

// Rules is a collection of rules for a valid go type
type Rules struct {
	TypeName string  `xml:"type,attr"`
	Children []*Rule `xml:"rule"`
}

// sort rules by priority
func (r *Rules) Len() int {
	return len(r.Children)
}
func (r *Rules) Swap(i, j int) {
	r.Children[i], r.Children[j] = r.Children[j], r.Children[i]
}
func (r *Rules) Less(i, j int) bool {
	return r.Children[i].Priority < r.Children[j].Priority
}

func (r *Rules) hasType(typ string) bool {
	return r.TypeName == typ
}

// RuleResult contains result of single rule expression
type RuleResult struct {
	name string      // rule name
	val  interface{} // rule result value
}

// Name returns the rule name
func (r *RuleResult) Name() string {
	return r.name
}

// Val returns the value of the result as an interface{}
// When resultType="CustomType", assert:  r.Val().(packageName.CustomType), for getting the concrete type
func (r *RuleResult) Val() interface{} {
	return r.val
}

// StringVal asserts value of the result for string type
func (r *RuleResult) StringVal() string {
	return r.val.(string)
}

// BoolVal asserts value of the result for bool type
func (r *RuleResult) BoolVal() bool {
	return r.val.(bool)
}

// FloatVal asserts value of the result for float64 type
func (r *RuleResult) FloatVal() float64 {
	return r.val.(float64)
}

// FloatArrayVal asserts value of the result for []float64 type
func (r *RuleResult) FloatArrayVal() []float64 {
	return r.val.([]float64)
}

// StringArrayVal asserts value of the result for []string type
func (r *RuleResult) StringArrayVal() []string {
	return r.val.([]string)
}

// MapVal asserts value of the result for map[string]interface{} type
func (r *RuleResult) MapVal() map[string]interface{} {
	return r.val.(map[string]interface{})
}

// Parser holds the rules from a rule file
type Parser struct {
	Name     xml.Name `xml:"roulette"`
	Children []*Rules `xml:"rules"`

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
			//fmt.Printf("%+v\n", rule)
			rule.compile(p.delimLeft, p.delimRight)
		}
		// sort by rule priority
		sort.Sort(p.sortedChildren[typeName])

	}
}

// ResultOne returns the top priority rule's result for the val's type.
func (p *Parser) ResultOne(val interface{}) (*RuleResult, error) {

	res := p.results(val, 1)
	if len(res) == 0 {
		return nil, errors.New("No rule was triggered!")
	}
	return res[0], nil

}

// ResultAll returns all rule results for the val's type sorted by priority
func (p *Parser) ResultAll(val interface{}) []*RuleResult {
	return p.results(val, len(p.Children))
}

// results runs rule expressions against a type and returns a slice of RuleResult
func (p *Parser) results(val interface{}, end int) []*RuleResult {

	var resultsArr []*RuleResult
	typeName := structs.Name(val)

	for _, rule := range p.sortedChildren[typeName].Children[:end] {

		ruleResult := &RuleResult{name: rule.Name}

		data := make(map[string]interface{})
		data[structs.Name(val)] = val

		var result bytes.Buffer
		// execute rules
		err := rule.Template.Execute(&result, data)
		if err != nil {
			log.Println(err)
		}

		switch rule.ResultType {
		case "string":
			var stringVar bool
			err := json.Unmarshal(result.Bytes(), &stringVar)
			if err != nil {
				continue
			}
			ruleResult.val = stringVar
			break
		case "bool":
			var boolVar bool
			err := json.Unmarshal(result.Bytes(), &boolVar)
			if err != nil {
				continue
			}
			ruleResult.val = boolVar
			break

		case "float":
			var floatVar float64
			err := json.Unmarshal(result.Bytes(), &floatVar)
			if err != nil {
				continue
			}
			ruleResult.val = floatVar
			break

		case "float_array":
			var sliceVar []float64
			err := json.Unmarshal(result.Bytes(), &sliceVar)
			if err != nil {
				continue
			}
			ruleResult.val = sliceVar
			break
		case "string_array":
			var sliceVar []string
			err := json.Unmarshal(result.Bytes(), &sliceVar)
			if err != nil {
				continue
			}
			ruleResult.val = sliceVar
			break

		case "map":
			var valMap map[string]interface{}
			err := json.Unmarshal(result.Bytes(), &valMap)
			if err != nil {
				continue
			}
			ruleResult.val = valMap
			break

		case structs.Name(val):
			newVal := reflect.New(reflect.TypeOf(val))
			res := bytes.TrimSpace(result.Bytes())
			err := json.Unmarshal(res, newVal.Interface())
			if err != nil {
				log.Println(err)
				continue
			}
			ruleResult.val = newVal.Interface()
			break

		default:
			var valInterface interface{}
			err = json.Unmarshal(result.Bytes(), &valInterface)
			if err != nil {
				continue
			}

			ruleResult.val = valInterface

		}

		resultsArr = append(resultsArr, ruleResult)

	}

	return resultsArr
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
	}

	err := xml.Unmarshal(data, parser)
	if err != nil {
		return nil, err
	}

	parser.compile()

	return parser, nil
}
