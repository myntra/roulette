package roulette

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"text/template"

	"github.com/fatih/structs"
	"github.com/ulule/deepcopier"
)

// http://stackoverflow.com/questions/6395076/in-golang-using-reflect-how-do-you-set-the-value-of-a-struct-field

func within(fieldVal int, minVal int, maxVal int, prevVal ...string) string {
	if len(prevVal) > 0 {
		if prevVal[0] == "false" {
			return "false"
		}
	}

	//fmt.Println(typeName, fieldVal, minVal, maxVal)
	if fieldVal >= minVal && fieldVal <= maxVal {
		return "true"
	}
	return "false"
}

func gte(fieldVal int, minVal int, prevVal ...string) string {

	if len(prevVal) > 0 {
		if prevVal[0] == "false" {
			return "false"
		}
	}

	//fmt.Println(typeName, fieldVal, maxVal)
	if fieldVal >= minVal {
		return "true"
	}
	return "false"
}

func lte(fieldVal int, maxVal int, prevVal ...string) string {

	if len(prevVal) > 0 {
		if prevVal[0] == "false" {
			return "false"
		}
	}

	//fmt.Println(typeName, fieldVal, maxVal)
	if fieldVal <= maxVal {
		return "true"
	}
	return "false"
}

func eql(fieldVal interface{}, val interface{}, prevVal ...string) string {
	if len(prevVal) > 0 {
		if prevVal[0] == "false" {
			return "false"
		}
	}

	if fieldVal == val {
		return "true"
	}
	return "false"
}

// set sets a field in a type
func set(typeVal interface{}, fieldTypeVal string, val interface{}, prevVal ...string) string {
	if len(prevVal) > 0 {
		if prevVal[0] == "false" {
			return ""
		}
	}

	// get map
	m := structs.Map(typeVal)
	m[fieldTypeVal] = val

	// get new type
	newTypeVal := reflect.New(reflect.TypeOf(typeVal))

	deepcopier.Copy(typeVal).To(newTypeVal.Interface())
	setFieldIn(newTypeVal.Interface(), fieldTypeVal, val)

	b, err := json.Marshal(newTypeVal.Interface())
	if err != nil {
		log.Println(err)
		return ""
	}

	ret := string(b)

	//fmt.Println(ret)

	return ret
}

var defaultFuncMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"within": within,
	"gte":    gte,
	"lte":    lte,
	"eql":    eql,
	"set":    set,
}

func setFieldIn(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("Provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}
