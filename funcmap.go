package roulette

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"text/template"
	"unicode"
)

// code lifted from release-branch.go1.8/src/text/template/funcs.go

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

// indirectInterface returns the concrete value in an interface value,
// or else the zero reflect.Value.
// That is, if v represents the interface value x, the result is the same as reflect.ValueOf(x):
// the fact that x was an interface value is forgotten.
func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}

// goodFunc reports whether the function or method has the right result signature.
func goodFunc(typ reflect.Type) bool {
	// We allow functions with 1 result or 2 results where the second is an error.
	switch {
	case typ.NumOut() == 1:
		return true
	case typ.NumOut() == 2 && typ.Out(1) == errorType:
		return true
	}
	return false
}

// goodName reports whether the function name is a valid identifier.
func goodName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_':
		case i == 0 && !unicode.IsLetter(r):
			return false
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			return false
		}
	}
	return true
}

// IsTrue reports whether the value is 'true', in the sense of not the zero of its type,
// and whether the value has a meaningful truth value. This is the definition of
// truth used by if and other such actions.
func IsTrue(val interface{}) (truth, ok bool) {
	return isTrue(reflect.ValueOf(val))
}

func isTrue(val reflect.Value) (truth, ok bool) {
	if !val.IsValid() {
		// Something like var x interface{}, never set. It's a form of nil.
		return false, true
	}
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		truth = val.Len() > 0
	case reflect.Bool:
		truth = val.Bool()
	case reflect.Complex64, reflect.Complex128:
		truth = val.Complex() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		truth = !val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		truth = val.Int() != 0
	case reflect.Float32, reflect.Float64:
		truth = val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		truth = val.Uint() != 0
	case reflect.Struct:
		truth = true // Struct values are always true.
	default:
		return
	}
	return truth, true
}

func truth(arg reflect.Value) bool {
	t, _ := isTrue(indirectInterface(arg))
	return t
}

// Comparison.

// TODO: Perhaps allow comparison between signed and unsigned integers.

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	errBadComparison     = errors.New("incompatible types for comparison")
	errNoComparison      = errors.New("missing argument for comparison")
)

type kind int

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	stringKind
	uintKind
)

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
}

// eq evaluates the comparison a == b || a == c || ...
func eq(arg1 reflect.Value, arg2 ...reflect.Value) (bool, error) {
	for i := range arg2 {
		prevArg := arg2[i]
		if !truth(prevArg) {
			return false, nil
		}
	}
	v1 := indirectInterface(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	if len(arg2) == 0 {
		return false, errNoComparison
	}
	for _, arg := range arg2 {
		v2 := indirectInterface(arg)
		k2, err := basicKind(v2)
		if err != nil {
			return false, err
		}
		truth := false
		if k1 != k2 {
			// Special case: Can compare integer values regardless of type's sign.
			switch {
			case k1 == intKind && k2 == uintKind:
				truth = v1.Int() >= 0 && uint64(v1.Int()) == v2.Uint()
			case k1 == uintKind && k2 == intKind:
				truth = v2.Int() >= 0 && v1.Uint() == uint64(v2.Int())
			default:
				return false, errBadComparison
			}
		} else {
			switch k1 {
			case boolKind:
				truth = v1.Bool() == v2.Bool()
			case complexKind:
				truth = v1.Complex() == v2.Complex()
			case floatKind:
				truth = v1.Float() == v2.Float()
			case intKind:
				truth = v1.Int() == v2.Int()
			case stringKind:
				truth = v1.String() == v2.String()
			case uintKind:
				truth = v1.Uint() == v2.Uint()
			default:
				panic("invalid kind")
			}
		}
		if truth {
			return true, nil
		}
	}
	return false, nil
}

// ne evaluates the comparison a != b.
func ne(arg1, arg2 reflect.Value, arg3 ...reflect.Value) (bool, error) {
	// != is the inverse of ==.
	if len(arg3) > 0 {
		if !truth(arg3[0]) {
			return false, nil
		}
	}
	equal, err := eq(arg1, arg2)
	return !equal, err
}

// lt evaluates the comparison a < b.
func lt(arg1, arg2 reflect.Value, arg3 ...reflect.Value) (bool, error) {
	//fmt.Println("lt", arg1, arg2)
	if len(arg3) > 0 {
		if !truth(arg3[0]) {
			return false, nil
		}
	}

	v1 := indirectInterface(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	v2 := indirectInterface(arg2)
	k2, err := basicKind(v2)
	if err != nil {
		return false, err
	}
	truth := false
	if k1 != k2 {
		// Special case: Can compare integer values regardless of type's sign.
		switch {
		case k1 == intKind && k2 == uintKind:
			truth = v1.Int() < 0 || uint64(v1.Int()) < v2.Uint()
		case k1 == uintKind && k2 == intKind:
			truth = v2.Int() >= 0 && v1.Uint() < uint64(v2.Int())
		default:
			return false, errBadComparison
		}
	} else {
		switch k1 {
		case boolKind, complexKind:
			return false, errBadComparisonType
		case floatKind:
			truth = v1.Float() < v2.Float()
		case intKind:
			truth = v1.Int() < v2.Int()
		case stringKind:
			truth = v1.String() < v2.String()
		case uintKind:
			truth = v1.Uint() < v2.Uint()
		default:
			panic("invalid kind")
		}
	}

	//fmt.Println("lt", truth)
	return truth, nil
}

// le evaluates the comparison <= b.
func le(arg1, arg2 reflect.Value, arg3 ...reflect.Value) (bool, error) {
	if len(arg3) > 0 {
		if !truth(arg3[0]) {
			return false, nil
		}
	}
	// <= is < or ==.
	lessThan, err := lt(arg1, arg2)
	if lessThan || err != nil {
		return lessThan, err
	}
	return eq(arg1, arg2)
}

// gt evaluates the comparison a > b.
func gt(arg1, arg2 reflect.Value, arg3 ...reflect.Value) (bool, error) {
	if len(arg3) > 0 {
		if !truth(arg3[0]) {
			return false, nil
		}
	}

	// > is the inverse of <=.
	lessOrEqual, err := le(arg1, arg2)
	if err != nil {
		return false, err
	}
	return !lessOrEqual, nil
}

// ge evaluates the comparison a >= b.
func ge(arg1, arg2 reflect.Value, arg3 ...reflect.Value) (bool, error) {
	if len(arg3) > 0 {
		if !truth(arg3[0]) {
			return false, nil
		}
	}
	// >= is the inverse of <.
	lessThan, err := lt(arg1, arg2)
	if err != nil {
		return false, err
	}

	return !lessThan, nil
}

// not returns the Boolean negation of its argument.
func not(arg0 reflect.Value, arg1 ...reflect.Value) bool {
	if len(arg1) > 0 {
		if !truth(arg1[0]) {
			return false
		}
	}
	return !truth(arg0)
}

func and(arg0 reflect.Value, args ...reflect.Value) (bool, error) {

	if !truth(arg0) {
		return false, nil
	}

	for i := range args {
		arg0 = args[i]
		if !truth(arg0) {
			return false, nil
		}
	}
	return true, nil
}

func or(arg0, arg1 reflect.Value, args ...reflect.Value) (bool, error) {
	//fmt.Println("or ", truth(arg0), truth(arg1))
	for i := range args {
		prevArg := args[i]
		if !truth(prevArg) {
			return false, nil
		}
	}

	if truth(arg0) {
		return true, nil
	}

	if truth(arg1) {
		return true, nil
	}

	return false, nil
}

// within evaluates the comparison minVal <= a <= maxVal
func within(arg1, arg2, arg3 reflect.Value, arg4 ...reflect.Value) (bool, error) {
	if len(arg4) > 0 {
		if !truth(arg4[0]) {
			return false, nil
		}
	}

	greaterOrEqual, err := ge(arg1, arg2)
	if err != nil {
		return false, err
	}

	lessOrEqual, err := le(arg1, arg3)
	if err != nil {
		return false, err
	}

	return greaterOrEqual && lessOrEqual, nil
}

// ternary operator
func tern(cond bool, t, f interface{}) interface{} {
	if cond {
		return t
	}
	return f
}

// validateFuncs validates additional functions to be added to the parser
// Functions must be of the signature: f(arg1,arg2, prevVal ...bool)bool
// See funcmap.go for examples.
func validateFuncs(funcMap template.FuncMap) error {
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
	}

	return nil
}

// remove when the below functions are added to spring

// toFloat64 converts 64-bit floats
func toFloat64(v interface{}) float64 {
	if str, ok := v.(string); ok {
		iv, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}
		return iv
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return float64(val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return float64(val.Uint())
	case reflect.Uint, reflect.Uint64:
		return float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Bool:
		if val.Bool() == true {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func floor(a interface{}) float64 {
	aa := toFloat64(a)
	return math.Floor(aa)
}

func ceil(a interface{}) float64 {
	aa := toFloat64(a)
	return math.Ceil(aa)
}

func round(a interface{}, p int, r_opt ...float64) float64 {
	roundOn := .5
	if len(r_opt) > 0 {
		roundOn = r_opt[0]
	}
	val := toFloat64(a)
	places := toFloat64(p)

	var round float64
	pow := math.Pow(10, places)
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}

var defaultFuncMap = template.FuncMap{
	"in": within,
	// Comparisons
	"eq":    eq, // ==
	"ge":    ge, // >=
	"gt":    gt, // >
	"le":    le, // <=
	"lt":    lt, // <
	"ne":    ne, // !=
	"not":   not,
	"and":   and,
	"or":    or,
	"tern":  tern,
	"ceil":  ceil,
	"floor": floor,
	"round": round,
}
