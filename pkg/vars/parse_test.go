package vars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func subTxt(n string) *SubTxt         { return &SubTxt{Val: n} }
func subVar(n, p, exp string) *SubVar { return &SubVar{Name: n, Params: p, Expr: exp} }

func TestParseExpr(t *testing.T) {
	assert.Equal(t, Subs{subTxt("values('"), subVar("random_int", "15-95", "@random_int(15-95)"),
		subTxt("','"), subVar("身份证", "", "@身份证"), subTxt("')")}, ParseExpr("values('@random_int(15-95)','@身份证')"))
	assert.Equal(t, Subs{subVar("中文", "", "@中文")}, ParseExpr("@中文"))
	assert.Equal(t, Subs{subVar("fn", "", "@fn")}, ParseExpr("@fn"))
	assert.Equal(t, Subs{subVar("fn.1", "", "@fn.1")}, ParseExpr("@fn.1"))
	assert.Equal(t, Subs{subVar("fn-1", "", "@fn-1")}, ParseExpr("@fn-1"))
	assert.Equal(t, Subs{subVar("fn_1", "", "@fn_1")}, ParseExpr("@fn_1"))
	assert.Equal(t, Subs{subVar("fn", "", "@fn"), subTxt("@")}, ParseExpr("@fn@"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@{fn}")}, ParseExpr("abc@{fn}"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@<fn>")}, ParseExpr("abc@<fn>"))
	assert.Equal(t, Subs{subVar("fn", "", "@fn"), subVar("fn", "", "@fn")}, ParseExpr("@fn@fn"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@fn"), subVar("fn", "", "@{fn}"), subTxt("efg")}, ParseExpr("abc@fn@{fn}efg"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@fn"), subVar("fn", "1", "@{fn(1)}"), subTxt("efg")}, ParseExpr("abc@fn@{fn(1)}efg"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn", "", "@fn"), subVar("中文", "1", "@{中文(1)}"), subTxt("efg")}, ParseExpr("abc@fn@{中文(1)}efg"))
	assert.Equal(t, Subs{subVar("fn", "100", "@fn(100)")}, ParseExpr("@fn(100)"))
	assert.Equal(t, Subs{subTxt("@")}, ParseExpr("@"))
	assert.Equal(t, Subs{subTxt("@@")}, ParseExpr("@@"))
}
