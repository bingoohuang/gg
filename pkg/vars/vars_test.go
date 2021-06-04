package vars

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVars(t *testing.T) {
	m := map[string]func() GenFn{
		"name": func() GenFn { return func() interface{} { return "bingoo" } },
	}
	mv := NewMapGenValue(m)
	s := Eval("hello {name}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, mv.Vars)

	s = Eval("hello {{name}}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, mv.Vars)

	s = Eval("hello ${name}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, mv.Vars)

	mv = NewMapGenValue(map[string]func() GenFn{})
	s = Eval("hello ${name}", mv)
	assert.Equal(t, "hello name", s)
	assert.Equal(t, map[string]bool{"name": true}, mv.MissedVars)
}
