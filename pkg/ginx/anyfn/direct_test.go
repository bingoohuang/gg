package anyfn_test

import (
	"errors"
	"testing"

	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDirect(t *testing.T) {
	af := anyfn.NewAdapter()
	r := adapt.Adapt(gin.New(), af)

	r.GET("/direct1", af.F(func() interface{} {
		return anyfn.DirectResponse{Code: 203}
	}))
	r.GET("/direct2", af.F(func() interface{} {
		return &anyfn.DirectResponse{Error: errors.New("abc")}
	}))
	r.GET("/direct3", af.F(func() interface{} {
		return &anyfn.DirectResponse{String: "ABC"}
	}))
	r.GET("/direct4", af.F(func() interface{} {
		return &anyfn.DirectResponse{
			JSON: struct {
				Name string `json:"name"`
			}{
				Name: "ABC",
			},
			Header: map[string]string{"Xx-Server": "DDD"},
		}
	}))

	rr := gintest.Get("/direct1", r)
	assert.Equal(t, 203, rr.StatusCode())
	assert.Equal(t, "", rr.Body())

	rr = gintest.Get("/direct2", r)
	assert.Equal(t, 500, rr.StatusCode())
	assert.Equal(t, "abc", rr.Body())

	rr = gintest.Get("/direct3", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, `ABC`, rr.Body())

	rr = gintest.Get("/direct4", r)
	assert.Equal(t, 200, rr.StatusCode())
	assert.Equal(t, `{"name":"ABC"}`, rr.Body())
	assert.Equal(t, `DDD`, rr.Header()["Xx-Server"][0])
}
