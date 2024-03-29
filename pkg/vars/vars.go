package vars

import (
	"fmt"
	"regexp"
	"strings"
)

type GenFn func() interface{}

type MapGenValue struct {
	GenMap     map[string]func(params string) GenFn
	Map        map[string]GenFn
	MissedVars map[string]bool
	Vars       map[string]interface{}
}

func NewMapGenValue(m map[string]func(params string) GenFn) *MapGenValue {
	return &MapGenValue{
		GenMap:     m,
		Map:        make(map[string]GenFn),
		Vars:       make(map[string]interface{}),
		MissedVars: make(map[string]bool),
	}
}

func (m *MapGenValue) Value(name, params, expr string) interface{} {
	return m.GetValue(name, params, expr)
}

func (m *MapGenValue) GetValue(name, params, expr string) interface{} {
	if fn, ok := m.Map[name]; ok {
		return fn()
	}

	var f GenFn

	if fn, ok := m.GenMap[name]; ok {
		ff := fn(params)
		f = func() interface{} {
			v := ff()
			m.Vars[name] = v
			return v
		}
	} else {
		f = func() interface{} { return expr }
		m.MissedVars[name] = true
	}

	m.Map[name] = f
	return f()
}

type VarValue interface {
	GetValue(name, params, expr string) interface{}
}

type VarValueHandler func(name, params, expr string) interface{}

func (v VarValueHandler) GetValue(name, params, expr string) interface{} {
	return v(name, params, expr)
}

func EvalSubstitute(s string, varValue VarValue) string {
	return ParseSubstitute(s).Eval(varValue)
}

type Part interface {
	Eval(varValue VarValue) string
}

type Var struct {
	Name string
	Expr string
}

type Literal struct{ V string }

func (l Literal) Eval(VarValue) string { return l.V }
func (l Var) Eval(varValue VarValue) string {
	return fmt.Sprintf("%s", varValue.GetValue(l.Name, "", l.Expr))
}

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

		parts = append(parts, &Var{Name: vn, Expr: sub})
	}

	if start < len(s) {
		parts = append(parts, &Literal{V: s[start:]})
	}

	return parts
}
