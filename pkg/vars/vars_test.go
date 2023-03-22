package vars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVars(t *testing.T) {
	m := map[string]func(params string) GenFn{
		"name": func(params string) GenFn { return func() interface{} { return "bingoo" } },
	}
	mv := NewMapGenValue(m)
	s := EvalSubstitute("hello {name}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, mv.Vars)

	s = EvalSubstitute("hello {{name}}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, mv.Vars)

	s = EvalSubstitute("hello ${name}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, mv.Vars)

	mv = NewMapGenValue(map[string]func(params string) GenFn{})
	s = EvalSubstitute("hello ${name}", mv)
	assert.Equal(t, "hello name", s)
	assert.Equal(t, map[string]bool{"name": true}, mv.MissedVars)
}
