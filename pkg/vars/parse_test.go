package vars

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func subTxt(n string) *SubTxt     { return &SubTxt{Val: n} }
func subVar(n string) *SubVar     { return &SubVar{Name: n} }
func subVarP(n, p string) *SubVar { return &SubVar{Name: n, Params: p} }

func TestParseExpr(t *testing.T) {
	assert.Equal(t, Subs{subVar("中文")}, ParseExpr("@中文"))
	assert.Equal(t, Subs{subVar("fn")}, ParseExpr("@fn"))
	assert.Equal(t, Subs{subVar("fn"), subTxt("@")}, ParseExpr("@fn@"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn")}, ParseExpr("abc@{fn}"))
	assert.Equal(t, Subs{subVar("fn"), subVar("fn")}, ParseExpr("@fn@fn"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn"), subVar("fn"), subTxt("efg")}, ParseExpr("abc@fn@{fn}efg"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn"), subVarP("fn", "1"), subTxt("efg")}, ParseExpr("abc@fn@{fn(1)}efg"))
	assert.Equal(t, Subs{subTxt("abc"), subVar("fn"), subVarP("中文", "1"), subTxt("efg")}, ParseExpr("abc@fn@{中文(1)}efg"))
	assert.Equal(t, Subs{subVarP("fn", "100")}, ParseExpr("@fn(100)"))
	assert.Equal(t, Subs{subTxt("@")}, ParseExpr("@"))
	assert.Equal(t, Subs{subTxt("@@")}, ParseExpr("@@"))
}
