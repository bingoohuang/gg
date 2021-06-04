package vars

import (
	"fmt"
	"regexp"
	"strings"
)

func EvalVarsMap(s string, genFnMap map[string]func() GenFn, vars map[string]interface{}) Result {
	missedVars := map[string]struct{}{}
	if vars == nil {
		vars = map[string]interface{}{}
	}

	result := Parse(s, genFnMap, missedVars).Eval(vars)
	result.MissedVars = missedVars
	return result
}

func Eval(s string, genFnMap map[string]func() GenFn) Result {
	return EvalVarsMap(s, genFnMap, nil)
}

type Part interface {
	Eval(vars map[string]interface{}) string
}

type GenFn func() interface{}

type Var struct {
	Gen  GenFn
	Name string
}

type Literal struct{ V string }

func (l Literal) Eval(map[string]interface{}) string { return l.V }
func (l Var) Eval(vars map[string]interface{}) string {
	v, ok := vars[l.Name]
	if !ok {
		v = l.Gen()
		vars[l.Name] = v
	}

	return fmt.Sprintf("%s", v)
}

func (l Parts) Eval(vars map[string]interface{}) Result {
	if vars == nil {
		vars = map[string]interface{}{}
	}

	sb := strings.Builder{}
	for _, p := range l {
		sb.WriteString(p.Eval(vars))
	}
	return Result{
		Value: sb.String(),
		Vars:  vars,
	}
}

type Parts []Part

var varRe = regexp.MustCompile(`\$?\{[^{}]+?\}|\{\{[^{}]+?\}\}`)

type Result struct {
	Value      string
	Vars       map[string]interface{}
	MissedVars map[string]struct{}
}

func Parse(s string, genFnMap map[string]func() GenFn, missedVars map[string]struct{}) (parts Parts) {
	locs := varRe.FindAllStringSubmatchIndex(s, -1)
	start := 0

	localGenFn := make(map[string]GenFn)

	for _, loc := range locs {
		parts = append(parts, &Literal{V: s[start:loc[0]]})
		sub := s[loc[0]+1 : loc[1]-1]
		sub = strings.TrimPrefix(sub, "{")
		sub = strings.TrimSuffix(sub, "}")
		start = loc[1]

		vn := strings.ToLower(strings.TrimSpace(sub))
		if _, ok := localGenFn[vn]; !ok {
			if ff, fok := genFnMap[vn]; !fok {
				missedVars[vn] = struct{}{}
				localGenFn[vn] = func() interface{} { return sub }
			} else {
				localGenFn[vn] = ff()
			}
		}

		parts = append(parts, &Var{Name: vn, Gen: localGenFn[vn]})
	}

	if start < len(s) {
		parts = append(parts, &Literal{V: s[start:]})
	}

	return parts
}
