package roulette

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/myntra/roulette/log"
)

// Ruleset ...
type Ruleset interface {
	Execute(vals interface{})
}

type ruleConfig struct {
	delimLeft    string
	delimRight   string
	defaultfuncs template.FuncMap
	userfuncs    template.FuncMap
	allfuncs     template.FuncMap

	expectTypes   []string
	resultAllowed bool
	resultKey     string
}

// Rule is a single rule expression. A rule expression is a valid go text/template
type Rule struct {
	Name     string `xml:"name,attr"`
	Priority int    `xml:"priority,attr"`
	Expr     string `xml:",innerxml"`
	config   ruleConfig
}

func (r Rule) isValid(filterTypesArr []string) error {

	if strings.Contains(r.Expr, r.config.resultKey) && !r.config.resultAllowed {
		return fmt.Errorf("rule expression contains result func but no type Result interface was set")
	}

	err := fmt.Errorf("rule expression expected types %s, got %s", r.config.expectTypes, filterTypesArr)

	if r.config.expectTypes == nil || filterTypesArr == nil {
		return err
	}

	if len(filterTypesArr) == 1 && filterTypesArr[0] == "map[string]interface {}" {
		return nil
	}

	// less
	if len(filterTypesArr) < len(r.config.expectTypes) {
		return err
	}

	// equal to
	if len(filterTypesArr) == len(r.config.expectTypes) {
		for i := range r.config.expectTypes {
			if filterTypesArr[i] != r.config.expectTypes[i] {
				return err
			}
		}
	}

	// greater than
	// all expected types should be present in the template data.
	for _, expectedType := range r.config.expectTypes {
		j := sort.SearchStrings(filterTypesArr, expectedType)
		found := j < len(filterTypesArr) && filterTypesArr[j] == expectedType
		if !found {
			return err
		}
	}

	return nil
}

type textTemplateRulesetConfig struct {
	workflowPattern string
	result          Result
	filterTypesArr  []string
	regex           *regexp.Regexp
	isWildCard      bool
}

// TextTemplateRuleset is a collection of rules for a valid go type
type TextTemplateRuleset struct {
	Name            string `xml:"name,attr"`
	FilterTypes     string `xml:"filterTypes,attr"`
	FilterStrict    bool   `xml:"filterStrict,attr"`
	DataKey         string `xml:"dataKey,attr"`
	ResultKey       string `xml:"resultKey,attr"`
	Rules           []Rule `xml:"rule"`
	PrioritiesCount string `xml:"prioritiesCount,attr"`
	Workflow        string `xml:"workflow,attr"`
	config          textTemplateRulesetConfig
	buf             *bufferPool
}

// sort rules by priority
func (t TextTemplateRuleset) Len() int {
	return len(t.Rules)
}
func (t TextTemplateRuleset) Swap(i, j int) {
	t.Rules[i], t.Rules[j] = t.Rules[j], t.Rules[i]
}
func (t TextTemplateRuleset) Less(i, j int) bool {
	return t.Rules[i].Priority < t.Rules[j].Priority
}

func (t TextTemplateRuleset) isValidForTypes(filterTypesArr ...string) bool {

	if len(filterTypesArr) == 0 {
		return false
	}

	if len(t.config.filterTypesArr) != len(filterTypesArr) && t.FilterStrict {
		return false
	}

	// if not filterStrict, look for atleast one match
	// if filterStrict look for atleast one mismatch
	for i, v := range t.config.filterTypesArr {
		j := sort.SearchStrings(t.config.filterTypesArr, v)
		found := j < len(t.config.filterTypesArr) && t.config.filterTypesArr[i] == v
		if !found {
			if t.FilterStrict {
				return false
			}
		} else {

			// filtering is not strict and one atleast one match found.
			if !t.FilterStrict {
				return true
			}
		}
	}

	return false
}

func getTypes(vals interface{}) []string {
	var types []string
	switch vals.(type) {
	case []interface{}:
		for _, v := range vals.([]interface{}) {
			typeName := strings.Replace(reflect.TypeOf(v).String(), "*", "", -1)
			types = append(types, typeName)
		}

		break

	default:
		typeName := strings.Replace(reflect.TypeOf(vals).String(), "*", "", -1)
		types = append(types, typeName)
	}

	return types
}

func (t TextTemplateRuleset) getTemplateData(vals interface{}) map[string]interface{} {

	//fmt.Println("getTemplateData", reflect.TypeOf(vals))
	// flatten multiple types in template map so that they can be referred by
	// dataKey

	tmplData := make(map[string]interface{})
	valsData := make(map[string]interface{})
	// index array of same types
	typeArrayIndex := make(map[string]int)

	switch vals.(type) {
	case []interface{}:
		nestedMap := make(map[string]interface{})
		for i, val := range vals.([]interface{}) {

			switch val.(type) {
			case []string, []int, []int32, []int64, []bool, []float32, []float64, []interface{}:
				replacer := strings.NewReplacer("[]", "", "{}", "")
				typeName := replacer.Replace(reflect.TypeOf(val).String())
				valsData[typeName+"slice"+strconv.Itoa(i)] = val

				break

			case map[string]int, map[string]string, map[string]bool:
				replacer := strings.NewReplacer("[", "", "]", "", "{}", "")
				typeName := replacer.Replace(reflect.TypeOf(val).String())
				valsData[typeName+strconv.Itoa(i)] = val
				break
			case map[string]interface{}:
				valsData = val.(map[string]interface{})
				break

			case bool, int, int32, int64, float32, float64:
				typeName := reflect.TypeOf(val).String()
				valsData[typeName+strconv.Itoa(i)] = val
				break

			default:
				typeName := strings.Replace(reflect.TypeOf(val).String(), "*", "", -1)
				pkgPaths := strings.Split(typeName, ".")
				//fmt.Println("getTemplateData: case []interface{}: ", i, pkgPaths)

				pkgTypeName := pkgPaths[len(pkgPaths)-1]
				indexPkgTypeName := pkgTypeName

				_, ok := typeArrayIndex[pkgTypeName]
				if !ok {
					typeArrayIndex[pkgTypeName] = 0
					nestedMap[pkgTypeName] = val

				} else {
					typeArrayIndex[pkgTypeName]++
				}

				indexPkgTypeName = pkgTypeName + strconv.Itoa(typeArrayIndex[pkgTypeName])
				nestedMap[indexPkgTypeName] = val

				packagePath := ""
				for _, p := range pkgPaths[:len(pkgPaths)-1] {
					packagePath = packagePath + p
				}
				//	fmt.Println("packagePath", packagePath)
				valsData[packagePath] = nestedMap
			}

		}

		break
	default:
		typeName := strings.Replace(reflect.TypeOf(vals).String(), "*", "", -1)
		pkgPaths := strings.Split(typeName, ".")
		//fmt.Println("default", pkgPaths)
		valsData[pkgPaths[0]] = map[string]interface{}{
			pkgPaths[1]: vals,
		}
	}

	valsData[t.ResultKey] = t.config.result
	tmplData[t.DataKey] = valsData

	//fmt.Println("map", tmplData)

	return tmplData

}

func (t TextTemplateRuleset) getLimit() int {

	if t.PrioritiesCount == "all" || t.PrioritiesCount == "" {
		return len(t.Rules)
	}

	prioritiesCount, err := strconv.ParseInt(t.PrioritiesCount, 10, 32)
	if err != nil {
		return len(t.Rules)
	}

	return int(prioritiesCount)
}

// Execute ...
func (t TextTemplateRuleset) Execute(vals interface{}) {

	if len(t.config.workflowPattern) > 0 && len(t.Workflow) > 0 {
		// if a regex is a no match
		if !t.config.regex.MatchString(t.config.workflowPattern) && !t.config.isWildCard {
			log.Warnf("ruleset %s is not valid for the current parser %s %s", t.Name, t.Workflow, t.config.workflowPattern)
			return
		}

		// lets check wildcard
		if !wildcardMatcher(t.Workflow, t.config.workflowPattern) {
			log.Warnf("ruleset %s is not valid for the current parser %s %s", t.Name, t.Workflow, t.config.workflowPattern)
			return
		}
	}

	types := getTypes(vals)
	sort.Strings(types)
	if !t.isValidForTypes(types...) {
		log.Warnf("invalid types %s skipping ruleset %s", types, t.Name)
		return
	}

	//	fmt.Println("types:", types)

	tmplData := t.getTemplateData(vals)

	successCount := 0
	limit := t.getLimit()

	for _, rule := range t.Rules {

		// validate if one of the types exist in the expression.
		// log.Printf("test rule %s", rule.Name)

		err := rule.isValid(types)
		if err != nil {
			log.Warnf("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		tmpl, err := template.
			New(rule.Name).Delims(
			rule.config.delimLeft, rule.config.delimRight).
			Funcs(rule.config.allfuncs).
			Parse(rule.Expr)

		if err != nil {
			log.Warnf("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		buf := t.buf.get()
		defer t.buf.put(buf)

		err = tmpl.Execute(buf, tmplData)
		if err != nil {
			log.Warnf("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		log.Infof("matched rule %s", rule.Name)

		var result bool
		err = json.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			log.Error("marhsal result error", err, buf.String(), tmplData)
			continue
		}

		// n high priority rules successful, break
		if result {
			successCount++
			if successCount == limit {
				break
			}
		}

	}
}
