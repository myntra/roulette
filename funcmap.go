package roulette

import "text/template"

func checkPrevVal(prevVal []bool) bool {
	if len(prevVal) > 0 {
		if !prevVal[0] {
			return false
		}
	}
	return true
}

func within(fieldVal int, minVal int, maxVal int, prevVal ...bool) bool {
	if !checkPrevVal(prevVal) {
		return false
	}

	if fieldVal >= minVal && fieldVal <= maxVal {
		return true
	}
	return false
}

func gte(fieldVal int, minVal int, prevVal ...bool) bool {
	if !checkPrevVal(prevVal) {
		return false
	}
	if fieldVal >= minVal {
		return true
	}
	return false
}

func lte(fieldVal int, maxVal int, prevVal ...bool) bool {
	if !checkPrevVal(prevVal) {
		return false
	}
	if fieldVal <= maxVal {
		return true
	}
	return false
}

func eql(fieldVal interface{}, val interface{}, prevVal ...bool) bool {
	if !checkPrevVal(prevVal) {
		return false
	}
	if fieldVal == val {

		return true
	}

	return false
}

var defaultFuncMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"within": within,
	"gte":    gte,
	"lte":    lte,
	"eql":    eql,
}
