package roulette

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// Ruleset ...
type Ruleset interface {
	Execute(vals interface{})
	Compile(left, right string, defaultfuncs, userfuncs template.FuncMap) error
	Result(result Result)
	Sort()
}

// Rule is a single rule expression. A rule expression is a valid go text/template
type Rule struct {
	Name     string `xml:"name,attr"`
	Priority int    `xml:"priority,attr"`
	Expr     string `xml:",innerxml"`

	delimLeft    string
	delimRight   string
	defaultfuncs template.FuncMap
	userfuncs    template.FuncMap
	allfuncs     template.FuncMap

	expectTypes   []string
	resultAllowed bool
	resultKey     string
	templateError error
}

// compile initialises rule templates
func (r *Rule) compile() {

	r.allfuncs = template.FuncMap{}

	sort.Strings(r.expectTypes)

	for k, v := range r.defaultfuncs {
		r.allfuncs[k] = v
	}
	for k, v := range r.userfuncs {
		r.allfuncs[k] = v
	}

	// remove all new lines from the expression
	r.Expr = strings.Replace(r.Expr, "\n", "", -1)

	return
}

func (r *Rule) isValid(filterTypesArr []string) error {

	if strings.Contains(r.Expr, r.resultKey) && !r.resultAllowed {
		return fmt.Errorf("rule expression contains result func but no type Result interface was set")
	}

	err := fmt.Errorf("rule expression expected types %s, got %s", r.expectTypes, filterTypesArr)

	if r.expectTypes == nil || filterTypesArr == nil {
		return err
	}

	// less
	if len(filterTypesArr) < len(r.expectTypes) {
		return err
	}

	// equal to
	if len(filterTypesArr) == len(r.expectTypes) {
		for i := range r.expectTypes {
			if filterTypesArr[i] != r.expectTypes[i] {
				return err
			}
		}
	}

	// greater than
	// all expected types should be present in the template data.
	for _, expectedType := range r.expectTypes {
		j := sort.SearchStrings(filterTypesArr, expectedType)
		found := j < len(filterTypesArr) && filterTypesArr[j] == expectedType
		if !found {
			return err
		}
	}

	return nil
}

// TextTemplateRuleset is a collection of rules for a valid go type
type TextTemplateRuleset struct {
	Name            string  `xml:"name,attr"`
	FilterTypes     string  `xml:"filterTypes,attr"`
	FilterStrict    bool    `xml:"filterStrict,attr"`
	DataKey         string  `xml:"dataKey,attr"`
	ResultKey       string  `xml:"resultKey,attr"`
	Rules           []*Rule `xml:"rule"`
	PrioritiesCount string  `xml:"prioritiesCount"`

	result         Result
	filterTypesArr []string
}

// sort rules by priority
func (r *TextTemplateRuleset) Len() int {
	return len(r.Rules)
}
func (r *TextTemplateRuleset) Swap(i, j int) {
	r.Rules[i], r.Rules[j] = r.Rules[j], r.Rules[i]
}
func (r *TextTemplateRuleset) Less(i, j int) bool {
	return r.Rules[i].Priority < r.Rules[j].Priority
}

func (r *TextTemplateRuleset) isValidForTypes(filterTypesArr ...string) bool {

	if r.filterTypesArr == nil || filterTypesArr == nil {
		return false
	}

	if len(r.filterTypesArr) != len(filterTypesArr) && r.FilterStrict {
		return false
	}

	// if not filterStrict, look for atleast one match
	// if filterStrict look for atleast one mismatch
	for i, v := range r.filterTypesArr {
		j := sort.SearchStrings(r.filterTypesArr, v)
		found := j < len(r.filterTypesArr) && r.filterTypesArr[i] == v
		if !found {
			if r.FilterStrict {
				return false
			}
		} else {

			// filtering is not strict and one atleast one match found.
			if !r.FilterStrict {
				return true
			}
		}
	}

	return false
}

func (r *TextTemplateRuleset) getTypes(vals interface{}) []string {
	var types []string
	switch vals.(type) {
	case map[string]interface{}:
		for _, v := range vals.(map[string]interface{}) {
			typeName := strings.Replace(reflect.TypeOf(v).String(), "*", "", -1)
			types = append(types, typeName)
		}

		break
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

func (r *TextTemplateRuleset) getTemplateData(vals interface{}) map[string]interface{} {

	//fmt.Println("getTemplateData", reflect.TypeOf(vals))
	// flatten multiple types in template map so that they can be referred by
	// dataKey
	tmplData := make(map[string]interface{})
	valsData := make(map[string]interface{})

	switch vals.(type) {
	case map[string]interface{}:
		valsData = vals.(map[string]interface{})
		break
	case []interface{}:
		nestedMap := make(map[string]interface{})
		for i, val := range vals.([]interface{}) {

			switch val.(type) {
			case []string, []int32, []int64, []bool, []float32, []float64, []interface{}:
				replacer := strings.NewReplacer("[]", "", "{}", "")
				typeName := replacer.Replace(reflect.TypeOf(val).String())
				valsData[typeName+"slice"+strconv.Itoa(i)] = val

				break

			case map[string]int, map[string]string, map[string]bool, map[string]interface{}:
				replacer := strings.NewReplacer("[", "", "]", "", "{}", "")
				typeName := replacer.Replace(reflect.TypeOf(val).String())
				valsData[typeName+strconv.Itoa(i)] = val
				break

			case bool, int, int32, int64, float32, float64:
				typeName := reflect.TypeOf(val).String()
				valsData[typeName+strconv.Itoa(i)] = val
				break

			default:
				typeName := strings.Replace(reflect.TypeOf(val).String(), "*", "", -1)
				pkgPaths := strings.Split(typeName, ".")
				//fmt.Println("getTemplateData: case []interface{}: ", i, pkgPaths)

				nestedMap[pkgPaths[len(pkgPaths)-1]] = val
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

	valsData[r.ResultKey] = r.result
	tmplData[r.DataKey] = valsData

	//fmt.Println("map", tmplData)

	return tmplData

}

func (r *TextTemplateRuleset) getLimit() int {

	if r.PrioritiesCount == "all" || r.PrioritiesCount == "" {
		return len(r.Rules)
	}

	prioritiesCount, err := strconv.ParseInt(r.PrioritiesCount, 10, 32)
	if err != nil {
		return len(r.Rules)
	}

	return int(prioritiesCount)
}

// Compile ...
func (r *TextTemplateRuleset) Compile(left, right string, defaultfuncs, userfuncs template.FuncMap) error {
	if r.FilterTypes == "" {
		return fmt.Errorf("Missing required attribute filterTypes")
	}

	if r.DataKey == "" {
		return fmt.Errorf("Missing required attribute dataKey")
	}

	// replace spaces, commas
	replacer := strings.NewReplacer(" ", "", "*", " ")
	typeName := replacer.Replace(r.FilterTypes)
	r.filterTypesArr = strings.Split(typeName, ",")
	sort.Strings(r.filterTypesArr)
	//fmt.Println("filterTypesArr", r.filterTypesArr)
	// set expected types for a rule
	for _, rule := range r.Rules {
		for _, typeName := range r.filterTypesArr {
			if strings.Contains(rule.Expr, typeName) {
				rule.expectTypes = append(rule.expectTypes, typeName)
			}
		}
	}

	if r.ResultKey == "" {
		r.ResultKey = "result"
	}

	resultAllowed := true

	if r.result == nil {
		resultAllowed = false
	}

	for _, rule := range r.Rules {

		rule.resultAllowed = resultAllowed
		rule.resultKey = r.ResultKey
		rule.delimLeft = delimLeft
		rule.delimRight = delimRight
		rule.defaultfuncs = defaultfuncs
		rule.userfuncs = userfuncs

		rule.compile()
	}

	return nil
}

// Result ...
func (r *TextTemplateRuleset) Result(result Result) {
	r.result = result
}

// Sort ...
func (r *TextTemplateRuleset) Sort() {
	sort.Sort(r)
}

// Execute ...
func (r *TextTemplateRuleset) Execute(vals interface{}) {

	types := r.getTypes(vals)
	sort.Strings(types)
	if !r.isValidForTypes(types...) {
		log.Println("invalid types, skpping...", types)
		return
	}

	//	fmt.Println("types:", types)

	tmplData := r.getTemplateData(vals)

	successCount := 0
	limit := r.getLimit()

	for _, rule := range r.Rules {

		// validate if one of the types exist in the expression.

		err := rule.isValid(types)
		if err != nil {
			log.Printf("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		t, err := template.New(rule.Name).Delims(rule.delimLeft, rule.delimRight).Funcs(rule.allfuncs).Parse(rule.Expr)
		if err != nil {
			log.Printf("invalid rule %s, error: %v", rule.Name, err)
			return
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, tmplData)
		if err != nil {
			log.Printf("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		var result bool
		err = json.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			log.Println(err)
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
