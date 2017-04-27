package roulette

import (
	"log"
	"strings"
	"text/template"
)

// default delimeters
var delimLeft = "<r>"
var delimRight = "</r>"

// Rule is a single rule expression. A rule expression is a valid go text/template
type Rule struct {
	Name     string `xml:"name,attr"`
	Priority int    `xml:"priority,attr"`
	Expr     string `xml:",innerxml"`
	Template *template.Template
}

// compile initialises rule templates
func (r *Rule) compile(left, right string, defaultfuncs, userfuncs template.FuncMap) {
	allFuncs := template.FuncMap{}
	for k, v := range defaultfuncs {
		allFuncs[k] = v
	}
	for k, v := range userfuncs {
		allFuncs[k] = v
	}

	// remove all new lines from the expression
	r.Expr = strings.Replace(r.Expr, "\n", "", -1)

	t, err := template.New(r.Name).Delims(left, right).Funcs(allFuncs).Parse(r.Expr)
	if err != nil {
		// custom funcs can be injected later.
		if strings.Contains(err.Error(), "not defined") {
			log.Println("skip compiling rule", r.Name, "error :", err)
		} else {
			log.Println(err)
			return
		}
	}

	r.Template = t
}

// Rules is a collection of rules for a valid go type
type Rules struct {
	TypeName string  `xml:"types,attr"`
	DataKey  string  `xml:"dataKey,attr"`
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
