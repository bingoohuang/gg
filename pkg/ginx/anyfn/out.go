package anyfn

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/bingoohuang/gg/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type OutSupport interface {
	OutSupport(v interface{}, vs []interface{}, c *gin.Context) (bool, error)
}

type OutSupportFn func(v interface{}, vs []interface{}, c *gin.Context) (bool, error)

func (o OutSupportFn) OutSupport(v interface{}, vs []interface{}, c *gin.Context) (bool, error) {
	return o(v, vs, c)
}

func DirectDealerSupport(v interface{}, vs []interface{}, c *gin.Context) (bool, error) {
	if dv, ok := v.(DirectDealer); ok {
		dv.Deal(c)
		return ok, nil
	}

	return false, nil
}

func ErrorSupport(v interface{}, vs []interface{}, c *gin.Context) (bool, error) {
	if dv, ok := v.(error); ok {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", dv))
		return ok, nil
	}

	return false, nil
}

func DefaultSupport(v0 interface{}, vs []interface{}, g *gin.Context) (bool, error) {
	if v0 == nil {
		return true, nil
	}

	switch reflect.Indirect(reflect.ValueOf(v0)).Kind() {
	case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice:
		g.Render(http.StatusOK, ginx.JSONRender{Data: v0})
	default:
		g.String(http.StatusOK, "%v", v0)
	}

	return true, nil
}
