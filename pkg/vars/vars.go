package vars

import (
	"fmt"
	"regexp"
	"strings"
)

type GenFn func() interface{}

type MapGenValue struct {
	GenMap     map[string]func() GenFn
	Map        map[string]GenFn
	MissedVars map[string]bool
	Vars       map[string]interface{}
}

func NewMapGenValue(m map[string]func() GenFn) *MapGenValue {
	return &MapGenValue{
		GenMap:     m,
		Map:        make(map[string]GenFn),
		Vars:       make(map[string]interface{}),
		MissedVars: make(map[string]bool),
	}
}

func (m *MapGenValue) Value(name, params string) interface{} {
	return m.GetValue(name)
}

func (m *MapGenValue) GetValue(name string) interface{} {
	if fn, ok := m.Map[name]; ok {
		return fn()
	}

	var f GenFn

	if fn, ok := m.GenMap[name]; ok {
		ff := fn()
		f = func() interface{} {
			v := ff()
			m.Vars[name] = v
			return v
		}
	} else {
		f = func() interface{} { return name }
		m.MissedVars[name] = true
	}

	m.Map[name] = f
	return f()
}

type VarValue interface {
	GetValue(name string) interface{}
}

func EvalSubstitute(s string, varValue VarValue) string {
	return ParseSubstitute(s).Eval(varValue)
}

type Part interface {
	Eval(varValue VarValue) string
}

type Var struct {
	Name string
}

type Literal struct{ V string }

func (l Literal) Eval(VarValue) string      { return l.V }
func (l Var) Eval(varValue VarValue) string { return fmt.Sprintf("%s", varValue.GetValue(l.Name)) }

func (l Parts) Eval(varValue VarValue) string {
	sb := strings.Builder{}
	for _, p := range l {
		sb.WriteString(p.Eval(varValue))
	}
	return sb.String()
}

type Parts []Part

var varRe = regexp.MustCompile(`\$?\{[^{}]+?\}|\{\{[^{}]+?\}\}`)

func ParseSubstitute(s string) (parts Parts) {
	locs := varRe.FindAllStringSubmatchIndex(s, -1)
	start := 0

	for _, loc := range locs {
		parts = append(parts, &Literal{V: s[start:loc[0]]})
		sub := s[loc[0]+1 : loc[1]-1]
		sub = strings.TrimPrefix(sub, "{")
		sub = strings.TrimSuffix(sub, "}")
		start = loc[1]

		vn := strings.TrimSpace(sub)

		parts = append(parts, &Var{Name: vn})
	}

	if start < len(s) {
		parts = append(parts, &Literal{V: s[start:]})
	}

	return parts
}
