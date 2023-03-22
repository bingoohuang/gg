package fn_test

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/fn"
	"github.com/stretchr/testify/assert"
)

func TestGetFuncName(t *testing.T) {
	assert.Equal(t, "github.com/bingoohuang/gg/pkg/fn.GetFuncName", fn.GetFuncName(fn.GetFuncName))
}
