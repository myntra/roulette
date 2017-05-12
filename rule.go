package roulette

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// Ruleset ...
type Ruleset interface {
	Execute(vals interface{}) error
}

type ruleConfig struct {
	delimLeft    string
	delimRight   string
	defaultfuncs template.FuncMap
	userfuncs    template.FuncMap
	allfuncs     template.FuncMap

	expectTypes    []string
	expectTypesErr error
	noResultFunc   bool

	template    *template.Template
	templateErr error
}

// Rule is a single rule expression. A rule expression is a valid go text/template
type Rule struct {
	Name     string `xml:"name,attr"`
	Priority int    `xml:"priority,attr"`
	Expr     string `xml:",innerxml"`
	config   ruleConfig
}

func (r Rule) hasType(typeName string) bool {
	j := sort.SearchStrings(r.config.expectTypes, typeName)
	return j < len(r.config.expectTypes) && r.config.expectTypes[j] == typeName
}

func (r Rule) isValid(vals interface{}) error {

	if r.config.expectTypes == nil || vals == nil {
		return r.config.expectTypesErr
	}

	switch vals.(type) {
	case []interface{}:
		var typeName string
		typedVals := vals.([]interface{})
		size := len(typedVals)

		if size == 0 || len(r.config.expectTypes) == 0 {
			return r.config.expectTypesErr
		}

		if size == 1 && typedVals[0] == "map[string]interface {}" {
			return nil
		}

		if size < len(r.config.expectTypes) {
			return r.config.expectTypesErr
		}

		foundCount := 0
		for _, v := range typedVals {
			if reflect.ValueOf(v).Kind() == reflect.Ptr || reflect.ValueOf(v).Kind() == reflect.Interface {
				typeName = reflect.TypeOf(v).Elem().String()
			} else {
				typeName = reflect.TypeOf(v).String()
			}

			if foundCount == len(r.config.expectTypes) {
				return nil
			}

			if r.hasType(typeName) {
				foundCount++
			}

		}

		if foundCount != len(r.config.expectTypes) {
			return r.config.expectTypesErr
		}

	default:
		typeName := reflect.TypeOf(vals).Elem().String()
		hasType := r.hasType(typeName)
		if !hasType {
			return r.config.expectTypesErr
		}
	}

	return nil
}

type textTemplateRulesetConfig struct {
	result         Result
	filterTypesArr []string
	workflowMatch  bool
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

	config        textTemplateRulesetConfig
	bytesBuf      *bytesPool
	mapBuf        *mapPool
	sameTypeIndex map[int]string
	limit         int
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

func (t TextTemplateRuleset) isValid(vals interface{}) bool {

	switch vals.(type) {
	case []interface{}:
		var typeName string
		typedVals := vals.([]interface{})
		size := len(typedVals)

		if size == 0 || len(t.config.filterTypesArr) == 0 {
			return false
		}

		if t.FilterStrict {
			if size < len(t.config.filterTypesArr) {
				return false
			}
		}

		for _, v := range vals.([]interface{}) {
			if reflect.ValueOf(v).Kind() == reflect.Ptr || reflect.ValueOf(v).Kind() == reflect.Interface {
				typeName = reflect.TypeOf(v).Elem().String()
			} else {
				typeName = reflect.TypeOf(v).String()
			}

			hasType := t.hasType(typeName)

			if t.FilterStrict && !hasType {
				return false
			}

			if hasType {
				return true
			}

		}

		break

	default:

		typeName := reflect.TypeOf(vals).Elem().String()
		hasType := t.hasType(typeName)
		if t.FilterStrict && !hasType {
			return false
		}

		if hasType {
			return true
		}

	}

	return false
}

func (t TextTemplateRuleset) hasType(typeName string) bool {
	j := sort.SearchStrings(t.config.filterTypesArr, typeName)
	return j < len(t.config.filterTypesArr) && t.config.filterTypesArr[j] == typeName
}

type templateData struct {
	sync.RWMutex
	data map[string]interface{}
}

func (td *templateData) Get(key string) interface{} {
	td.RLock()
	defer td.RUnlock()
	return td.data[key]
}

func (td *templateData) Set(key string, val interface{}) bool {
	td.Lock()
	defer td.Unlock()
	td.data[key] = val
	return true
}

func (t TextTemplateRuleset) getTemplateData(tmplData, userTmplData map[string]interface{}, vals interface{}) {

	//fmt.Println("getTemplateData", reflect.TypeOf(vals))
	// flatten multiple types in template map so that they can be referred by
	// dataKey

	valsData := t.mapBuf.get()
	defer t.mapBuf.put(valsData)
	// index array of same types
	typeArrayIndex := make(map[string]int)

	switch vals.(type) {
	case []interface{}:
		nestedMap := t.mapBuf.get()
		defer t.mapBuf.put(nestedMap)

		indexPkgTypeName := t.bytesBuf.get()
		defer t.bytesBuf.put(indexPkgTypeName)

		for i, val := range vals.([]interface{}) {

			switch val.(type) {
			case []string, []int, []int32, []int64, []bool, []float32, []float64, []interface{}:
				typ := reflect.TypeOf(val).String()
				typeName := strings.Trim(typ, "[]")
				typeName = strings.Trim(typ, "{}")
				valsData[typeName+"slice"+strconv.Itoa(i)] = val

				break

			case map[string]int, map[string]string, map[string]bool:
				typ := reflect.TypeOf(val).String()
				typeName := strings.Trim(typ, "[")
				typeName = strings.Trim(typ, "]")
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

				var pkgPath string
				var typeName string

				if reflect.ValueOf(val).Kind() == reflect.Ptr || reflect.ValueOf(val).Kind() == reflect.Interface {
					pkgTypeName := reflect.TypeOf(val).Elem().String()
					periodIndex := strings.Index(pkgTypeName, ".")
					typeName = pkgTypeName[periodIndex+1:]
					pkgPath = pkgTypeName[:periodIndex]

				} else {
					typeName = reflect.TypeOf(val).String()
				}

				indexPkgTypeName.WriteString(typeName)

				_, ok := typeArrayIndex[typeName]
				if !ok {
					typeArrayIndex[typeName] = 0
					nestedMap[typeName] = val

				} else {
					typeArrayIndex[typeName]++
				}

				nextIndex, ok := t.sameTypeIndex[typeArrayIndex[typeName]]
				if !ok {
					nextIndex = strconv.Itoa(typeArrayIndex[typeName])
					t.sameTypeIndex[typeArrayIndex[typeName]] = nextIndex
				}

				indexPkgTypeName.WriteString(nextIndex)
				key := indexPkgTypeName.String()

				nestedMap[key] = val
				valsData[pkgPath] = nestedMap

				indexPkgTypeName.Reset()
			}

		}

		break
	default:
		pkgTypeName := reflect.TypeOf(vals).Elem().String()
		periodIndex := strings.Index(pkgTypeName, ".")
		typeName := pkgTypeName[periodIndex+1:]
		pkgPath := pkgTypeName[:periodIndex]

		//fmt.Println("default", pkgPaths)
		valsData[pkgPath] = map[string]interface{}{
			typeName: vals,
		}
	}

	valsData[t.ResultKey] = t.config.result

	templateData := &templateData{data: userTmplData}
	valsData["R"] = templateData
	tmplData[t.DataKey] = valsData

	//fmt.Println("map", tmplData)

}

var mutex = &sync.RWMutex{}

// Execute ...
func (t TextTemplateRuleset) Execute(vals interface{}) error {

	if !t.config.workflowMatch {
		//log.Infof("ruleset %s is not valid for the current parser %s %s", t.Name, t.Workflow)
		return fmt.Errorf("ruleset %s is not valid for the current parser %s", t.Name, t.Workflow)
	}

	if !t.isValid(vals) {
		//	//log.Infof("invalid types %s skipping ruleset %s", types, t.Name)
		return fmt.Errorf("invalid types %s skipping ruleset", t.Name)
	}

	mutex.Lock()
	defer mutex.Unlock()
	//	fmt.Println("types:", types)
	tmplData := t.mapBuf.get()
	userTmplData := t.mapBuf.get()
	defer t.mapBuf.put(tmplData)
	defer t.mapBuf.put(userTmplData)
	t.getTemplateData(tmplData, userTmplData, vals)

	successCount := 0

	for i := range t.Rules {

		rule := t.Rules[i]

		////log.Infof("rule  %s", rule.Name)
		if rule.config.noResultFunc {
			//log.Infof("rule expression contains result func but no type Result interface was set %s", rule.Name)
			continue
		}

		// validate if one of the types exist in the expression.
		err := rule.isValid(vals)
		if err != nil {
			//log.Infof("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		if rule.config.templateErr != nil {
			//log.Infof("invalid rule %s, error: %v", rule.Name, rule.config.templateErr)
			continue
		}

		buf := t.bytesBuf.get()
		defer t.bytesBuf.put(buf)

		err = rule.config.template.Execute(buf, tmplData)
		if err != nil {
			//log.Infof("invalid rule %s, error: %v", rule.Name, err)
			continue
		}

		var result bool

		res := strings.TrimSpace(buf.String())

		result, err = strconv.ParseBool(res)
		if err != nil {
			////log.Infof("parse result error", err)
			continue
		}

		// n high priority rules successful, break
		if result {
			//log.Infof("rule passed %s", rule.Name)
			successCount++
			if successCount == t.limit {
				break
			}
		} else {
			//log.Infof("rule failed %s", rule.Name)
		}

	}

	return nil
}
