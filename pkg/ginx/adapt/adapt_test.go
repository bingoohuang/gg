package adapt_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/gintest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNoAdapt(t *testing.T) {
	r := gin.New()

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/user/:name", func(c *gin.Context) {
		c.Set("Xyz", "First")
	}, func(c *gin.Context) {
		name := c.Param("name")
		c.Header("Xyz", c.GetString("Xyz")+" Second")
		c.String(http.StatusOK, "Hello %s", name)
	})

	// r.Run(":8080")

	rr := gintest.Get("/user/bingoohuang", r)
	assert.Equal(t, "Hello bingoohuang", rr.Body())
	assert.Equal(t, "First Second", rr.Header().Get("Xyz"))
}

func TestAdapt(t *testing.T) {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(func(f func(string) string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(http.StatusOK, f(StringArg(c)))
		}
	})

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/user/:name", func(name string) string {
		return fmt.Sprintf("Hello %s", name)
	})

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/direct/:name", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello Direct %s", c.Param("name"))
	})

	// r.Run(":8080")

	rr := gintest.Get("/user/bingoohuang", r)
	assert.Equal(t, "Hello bingoohuang", rr.Body())

	rr = gintest.Get("/direct/bingoohuang", r)
	assert.Equal(t, "Hello Direct bingoohuang", rr.Body())
}

func TestGroup(t *testing.T) {
	r := adapt.Adapt(gin.New())

	r.RegisterAdapter(func(f func(string) string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(http.StatusOK, f(StringArg(c)))
		}
	})

	v1 := r.Group("/v1")
	v1.POST("/login", func(user string) string { return "Hello1 " + user })

	v2 := r.Group("/v2")
	v2.POST("/login", func(user string) string { return "Hello2 " + user })

	rr := gintest.Post("/v1/login", r, gintest.Query("user", "bingoohuang"))
	assert.Equal(t, "Hello1 bingoohuang", rr.Body())

	rr = gintest.Post("/v2/login", r, gintest.Query("user", "dingoohuang"))
	assert.Equal(t, "Hello2 dingoohuang", rr.Body())
}

func StringArg(c *gin.Context) string {
	if len(c.Params) == 1 {
		return c.Params[0].Value
	}

	if q := c.Request.URL.Query(); len(q) == 1 {
		for _, v := range q {
			return v[0]
		}
	}

	return ""
}
