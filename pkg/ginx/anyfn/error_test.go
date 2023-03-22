package anyfn_test

import (
	"errors"
	"testing"

	"github.com/bingoohuang/gg/pkg/ginx"
	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	af := anyfn.NewAdapter()
	r := adapt.Adapt(gin.New(), af)

	r.Any("/error", af.F(func() error { return errors.New("error occurred") }))
	r.GET("/ok", af.F(func() error { return nil }))
	r.GET("/url", af.F(func(c *gin.Context) (string, error) { return c.Request.URL.String(), nil }))

	rr := gintest.Get("/error", r)
	assert.Equal(t, 500, rr.StatusCode())
	assert.Equal(t, "error: error occurred", rr.Body())

	rr = gintest.Get("/ok", r)
	assert.Equal(t, 200, rr.StatusCode())

	rr = gintest.Get("/url", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, "/url", rr.Body())
}

func TestError2(t *testing.T) {
	af := anyfn.NewAdapter()

	type Resp struct {
		Code    int
		Message string
	}
	af.PrependOutSupport(anyfn.OutSupportFn(func(v interface{}, vs []interface{}, c *gin.Context) (bool, error) {
		if err, ok := v.(error); ok {
			c.Render(505, ginx.JSONRender{Data: Resp{Code: 500, Message: err.Error()}})
			return true, nil
		}

		return false, nil
	}))
	r := adapt.Adapt(gin.New(), af)

	r.Any("/error", af.F(func() error { return errors.New("error occurred") }))
	r.GET("/ok", af.F(func() error { return nil }))
	r.GET("/url", af.F(func(c *gin.Context) (string, error) { return c.Request.URL.String(), nil }))

	rr := gintest.Get("/error", r)
	assert.Equal(t, 505, rr.StatusCode())
	assert.Equal(t, `{"code":500,"message":"error occurred"}`, rr.Body())

	rr = gintest.Get("/ok", r)
	assert.Equal(t, 200, rr.StatusCode())

	rr = gintest.Get("/url", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, "/url", rr.Body())
}
