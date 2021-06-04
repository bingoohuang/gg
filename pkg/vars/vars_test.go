package vars

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVars(t *testing.T) {
	s := Eval("hello {name}", map[string]func() GenFn{
		"name": func() GenFn { return func() interface{} { return "bingoo" } },
	})
	assert.Equal(t, "hello bingoo", s.Value)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, s.Vars)

	s = Eval("hello {{name}}", map[string]func() GenFn{
		"name": func() GenFn { return func() interface{} { return "bingoo" } },
	})
	assert.Equal(t, "hello bingoo", s.Value)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, s.Vars)

	s = Eval("hello ${name}", map[string]func() GenFn{
		"name": func() GenFn { return func() interface{} { return "bingoo" } },
	})
	assert.Equal(t, "hello bingoo", s.Value)
	assert.Equal(t, map[string]interface{}{"name": "bingoo"}, s.Vars)

	s = Eval("hello ${name}", map[string]func() GenFn{})
	assert.Equal(t, "hello name", s.Value)
	assert.Equal(t, map[string]struct{}{"name": {}}, s.MissedVars)
}
