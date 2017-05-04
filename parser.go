package roulette

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

// default delimeters
var delimLeft = "<r>"
var delimRight = "</r>"

// Parser interface provides methods for the executor to update the tree, execute the values and get result.
type Parser interface {
	Execute(vals interface{})
	GetResult() Result
}

// XMLData contains the parsed roulette xml tree
type XMLData struct {
	Name     xml.Name              `xml:"roulette"`
	Rulesets []TextTemplateRuleset `xml:"ruleset"`
}

// TextTemplateParser holds the rules from a rule file
type TextTemplateParser struct {
	xml          XMLData
	config       TextTemplateParserConfig
	defaultFuncs template.FuncMap
}

// Execute executes the parser's rulesets
func (p TextTemplateParser) Execute(vals interface{}) {
	for i := range p.xml.Rulesets {
		p.xml.Rulesets[i].Execute(vals)
	}
}

// GetResult returns the parser's result.
func (p TextTemplateParser) GetResult() Result {
	return p.config.Result
}

// Compile compiles the parser's rulesets
func (p *TextTemplateParser) compile() error {

	var _ Ruleset = TextTemplateRuleset{}

	for i := range p.xml.Rulesets {

		if p.xml.Rulesets[i].FilterTypes == "" {
			return fmt.Errorf("Missing required attribute filterTypes")
		}

		if p.xml.Rulesets[i].DataKey == "" {
			return fmt.Errorf("Missing required attribute dataKey")
		}

		for i, rune := range p.xml.Rulesets[i].FilterTypes {
			if !unicode.IsLetter(rune) && i == 0 {
				return fmt.Errorf("First character of filterTypes is not a letter")
			}

			if i == 0 {
				break
			}
		}

		// split filter types
		replacer := strings.NewReplacer(" ", "", "*", " ")
		typeName := replacer.Replace(p.xml.Rulesets[i].FilterTypes)
		filterTypesArr := strings.Split(typeName, ",")
		sort.Strings(filterTypesArr)

		textTemplateRulesetConfig := textTemplateRulesetConfig{
			workflowPattern: p.config.WorkflowPattern,
			result:          p.config.Result,
			filterTypesArr:  filterTypesArr,
		}

		p.xml.Rulesets[i].config = textTemplateRulesetConfig

		if p.xml.Rulesets[i].ResultKey == "" {
			p.xml.Rulesets[i].ResultKey = "result"
		}

		resultAllowed := true

		if p.xml.Rulesets[i].config.result == nil {
			resultAllowed = false
		}

		// set rule config
		for j := range p.xml.Rulesets[i].Rules {

			ruleConfig := ruleConfig{
				resultAllowed: resultAllowed,
				resultKey:     p.xml.Rulesets[i].ResultKey,
				delimLeft:     p.config.DelimLeft,
				delimRight:    p.config.DelimRight,
				defaultfuncs:  p.defaultFuncs,
				userfuncs:     p.config.Userfuncs,
			}

			p.xml.Rulesets[i].Rules[j].config = ruleConfig

			// set expcted types for a rule
			for _, typeName := range p.xml.Rulesets[i].config.filterTypesArr {
				if strings.Contains(p.xml.Rulesets[i].Rules[j].Expr, typeName) {
					p.xml.Rulesets[i].Rules[j].config.expectTypes = append(p.xml.Rulesets[i].Rules[j].config.expectTypes, typeName)
				}
			}

			p.xml.Rulesets[i].Rules[j].config.allfuncs = template.FuncMap{}

			sort.Strings(p.xml.Rulesets[i].Rules[j].config.expectTypes)

			for k, v := range p.xml.Rulesets[i].Rules[j].config.defaultfuncs {
				p.xml.Rulesets[i].Rules[j].config.allfuncs[k] = v
			}
			for k, v := range p.xml.Rulesets[i].Rules[j].config.userfuncs {
				p.xml.Rulesets[i].Rules[j].config.allfuncs[k] = v
			}

			// remove all new lines from the expression
			p.xml.Rulesets[i].Rules[j].Expr = strings.Replace(p.xml.Rulesets[i].Rules[j].Expr, "\n", "", -1)

		}

		sort.Sort(p.xml.Rulesets[i])
	}

	return nil
}

// TextTemplateParserConfig sets the optional config for the TextTemplateParser
type TextTemplateParserConfig struct {
	Userfuncs       template.FuncMap
	DelimLeft       string
	DelimRight      string
	WorkflowPattern string // filter rulesets based on the pattern
	Result          Result
}

// NewTextTemplateParser returns a new roulette format xml parser.
func NewTextTemplateParser(data []byte, config TextTemplateParserConfig) (Parser, error) {

	if config.DelimLeft == "" {
		config.DelimLeft = delimLeft
		config.DelimRight = delimRight
	}

	if config.Userfuncs == nil {
		config.Userfuncs = template.FuncMap{}
	} else {
		err := validateFuncs(config.Userfuncs)
		if err != nil {
			return nil, err
		}
	}

	xmldata := XMLData{}

	err := xml.Unmarshal(data, &xmldata)
	if err != nil {
		return nil, err
	}

	parser := TextTemplateParser{
		config:       config,
		defaultFuncs: defaultFuncMap,
		xml:          xmldata,
	}

	// compile rulesets
	err = parser.compile()
	if err != nil {
		return nil, err
	}

	return parser, nil
}

// NewParser returns a TextTemplateParser with default config
func NewParser(data []byte, config ...TextTemplateParserConfig) (Parser, error) {
	cfg := TextTemplateParserConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return NewTextTemplateParser(data, cfg)
}
