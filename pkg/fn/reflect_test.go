package fn_test

import (
	"github.com/bingoohuang/gg/pkg/fn"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFuncName(t *testing.T) {
	assert.Equal(t, "github.com/bingoohuang/gg/pkg/fn.GetFuncName", fn.GetFuncName(fn.GetFuncName))
}
