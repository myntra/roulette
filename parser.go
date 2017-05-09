package roulette

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig"
	"github.com/myntra/roulette/log"
)

// default delimeters
var delimLeft = "<r>"
var delimRight = "</r>"

var defaultSameTypeIndex = map[int]string{
	1: "1",
	2: "2",
	3: "3",
	4: "4",
	5: "5",
}

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
	xml           XMLData
	config        TextTemplateParserConfig
	defaultFuncs  template.FuncMap
	bytesBuf      *bytesPool
	mapBuf        *mapPool
	sameTypeIndex map[int]string
}

// Execute executes the parser's rulesets
func (p TextTemplateParser) Execute(vals interface{}) {
	var err error
	for i := range p.xml.Rulesets {
		err = p.xml.Rulesets[i].Execute(vals)
		if err != nil {
			log.Warn(err)
		}

	}
}

// GetResult returns the parser's result.
func (p TextTemplateParser) GetResult() Result {
	return p.config.Result
}

// Compile compiles the parser's rulesets
func (p *TextTemplateParser) compile() error {
	replacer := strings.NewReplacer(" ", "", "*", " ")
	newLineReplacer := strings.NewReplacer("\n", "")
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

		// assign buffer pool
		p.xml.Rulesets[i].bytesBuf = p.bytesBuf
		p.xml.Rulesets[i].mapBuf = p.mapBuf
		p.xml.Rulesets[i].sameTypeIndex = p.sameTypeIndex

		// limit
		if p.xml.Rulesets[i].PrioritiesCount == "all" || p.xml.Rulesets[i].PrioritiesCount == "" {
			p.xml.Rulesets[i].limit = len(p.xml.Rulesets[i].Rules)
		} else {
			prioritiesCount, err := strconv.ParseInt(p.xml.Rulesets[i].PrioritiesCount, 10, 32)
			if err != nil {
				p.xml.Rulesets[i].limit = len(p.xml.Rulesets[i].Rules)
			} else {
				p.xml.Rulesets[i].limit = int(prioritiesCount)
			}
		}

		// split filter types

		typeName := replacer.Replace(p.xml.Rulesets[i].FilterTypes)
		filterTypesArr := strings.Split(typeName, ",")
		sort.Strings(filterTypesArr)

		regex := regexp.MustCompile(p.xml.Rulesets[i].Workflow)
		workflowMatch := true

		if len(p.config.WorkflowPattern) > 0 && len(p.xml.Rulesets[i].Workflow) > 0 {
			if p.config.IsWildcardWorkflowPattern {
				if !wildcardMatcher(p.xml.Rulesets[i].Workflow, p.config.WorkflowPattern) {
					workflowMatch = false
				}

			} else {
				// if a regex is a no match
				if !regex.MatchString(p.config.WorkflowPattern) {
					workflowMatch = false
				}
			}

		}

		textTemplateRulesetConfig := textTemplateRulesetConfig{
			result:         p.config.Result,
			filterTypesArr: filterTypesArr,
			workflowMatch:  workflowMatch,
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
				delimLeft:    p.config.DelimLeft,
				delimRight:   p.config.DelimRight,
				defaultfuncs: p.defaultFuncs,
				userfuncs:    p.config.Userfuncs,
			}

			p.xml.Rulesets[i].Rules[j].config = ruleConfig

			// remove all new lines from the expression
			p.xml.Rulesets[i].Rules[j].Expr = newLineReplacer.Replace(p.xml.Rulesets[i].Rules[j].Expr)

			if strings.Contains(p.xml.Rulesets[i].Rules[j].Expr, p.xml.Rulesets[i].ResultKey) && !resultAllowed {
				p.xml.Rulesets[i].Rules[j].config.noResultFunc = true
			}

			// set expcted types for a rule
			for _, typeName := range p.xml.Rulesets[i].config.filterTypesArr {
				if strings.Contains(p.xml.Rulesets[i].Rules[j].Expr, typeName) {
					p.xml.Rulesets[i].Rules[j].config.expectTypes = append(p.xml.Rulesets[i].Rules[j].config.expectTypes, typeName)
				}
			}

			p.xml.Rulesets[i].Rules[j].config.expectTypesErr = fmt.Errorf("rule expression expected types %s",
				p.xml.Rulesets[i].Rules[j].config.expectTypes)

			p.xml.Rulesets[i].Rules[j].config.allfuncs = template.FuncMap{}

			sort.Strings(p.xml.Rulesets[i].Rules[j].config.expectTypes)

			for k, v := range p.xml.Rulesets[i].Rules[j].config.defaultfuncs {
				p.xml.Rulesets[i].Rules[j].config.allfuncs[k] = v
			}

			// append Masterminds/sprig funcs

			for k, v := range sprig.FuncMap() {
				p.xml.Rulesets[i].Rules[j].config.allfuncs[k] = v
			}

			for k, v := range p.xml.Rulesets[i].Rules[j].config.userfuncs {
				p.xml.Rulesets[i].Rules[j].config.allfuncs[k] = v
			}

			tmpl, err := template.
				New(p.xml.Rulesets[i].Rules[j].Name).Delims(
				p.xml.Rulesets[i].Rules[j].config.delimLeft, p.xml.Rulesets[i].Rules[j].config.delimRight).
				Funcs(p.xml.Rulesets[i].Rules[j].config.allfuncs).
				Parse(p.xml.Rulesets[i].Rules[j].Expr)

			p.xml.Rulesets[i].Rules[j].config.template = tmpl
			p.xml.Rulesets[i].Rules[j].config.templateErr = err

		}

		sort.Sort(p.xml.Rulesets[i])
	}

	return nil
}

// TextTemplateParserConfig sets the optional config for the TextTemplateParser
type TextTemplateParserConfig struct {
	Userfuncs                 template.FuncMap
	DelimLeft                 string
	DelimRight                string
	WorkflowPattern           string // filter rulesets based on the pattern
	Result                    Result
	IsWildcardWorkflowPattern bool
	LogLevel                  string //info, debug, warn, error, fatal. default is info
	LogPath                   string //stdout, /path/to/file . default is stdout
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

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	if config.LogPath == "" {
		config.LogPath = "stdout"
	}

	xmldata := XMLData{}

	err := xml.Unmarshal(data, &xmldata)
	if err != nil {
		return nil, err
	}

	parser := TextTemplateParser{
		config:        config,
		defaultFuncs:  defaultFuncMap,
		xml:           xmldata,
		bytesBuf:      newBytesPool(),
		mapBuf:        newMapPool(),
		sameTypeIndex: defaultSameTypeIndex,
	}

	// compile rulesets
	err = parser.compile()
	if err != nil {
		return nil, err
	}

	log.Init(config.LogLevel, config.LogPath)

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
