package vars

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func subTxt(n string) *SubTxt         { return &SubTxt{Val: n} }
func subVar(n, p, exp string) *SubVar { return &SubVar{Name: n, Params: p, Expr: exp} }

func TestParseExpr(t *testing.T) {
	assert.Equal(t, Subs{subVar("中文", "", "@中文")}, ParseExpr("@中文"))
	assert.Equal(t, Subs{subVar("fn", "", "@fn")}, ParseExpr("@fn"))
	assert.Equal(t, Subs{subVar("fn", "", "@fn"), subTxt("@")}, ParseExpr("@fn@"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@{fn}")}, ParseExpr("abc@{fn}"))
	assert.Equal(t, Subs{subVar("fn", "", "@fn"), subVar("fn", "", "@fn")}, ParseExpr("@fn@fn"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@fn"), subVar("fn", "", "@{fn}"), subTxt("efg")}, ParseExpr("abc@fn@{fn}efg"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@fn"), subVar("fn", "1", "@{fn(1)}"), subTxt("efg")}, ParseExpr("abc@fn@{fn(1)}efg"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@fn"), subVar("中文", "1", "@{中文(1)}"), subTxt("efg")}, ParseExpr("abc@fn@{中文(1)}efg"))
	assert.Equal(t, Subs{subVar("fn", "100", "@fn(100)")}, ParseExpr("@fn(100)"))
	assert.Equal(t, Subs{subTxt("@")}, ParseExpr("@"))
	assert.Equal(t, Subs{subTxt("@@")}, ParseExpr("@@"))
}
