package anyfn_test

import (
	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/gintest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type AuthUser struct {
	Name string
}

func TestMiddleware(t *testing.T) {
	af := anyfn.NewAdapter()
	r := adapt.Adapt(gin.New(), af)

	r.Use(func(c *gin.Context) {
		c.Set("AuthUser", AuthUser{Name: "TestAuthUser"})
	})

	doTest(t, r, af)
}

func TestMiddlewarePtr(t *testing.T) {
	af := anyfn.NewAdapter()
	r := adapt.Adapt(gin.New(), af)

	r.Use(func(c *gin.Context) {
		c.Set("AuthUser", &AuthUser{Name: "TestAuthUser"})
	})

	doTest(t, r, af)
}

func TestInSupport(t *testing.T) {
	user := AuthUser{Name: "TestAuthUser"}
	af := anyfn.NewAdapter()
	af.PrependInSupport(anyfn.InSupportFn(func(argIn anyfn.ArgIn, argsIn []anyfn.ArgIn, c *gin.Context) (reflect.Value, error) {
		if argIn.Type == reflect.TypeOf(AuthUser{}) {
			return anyfn.ConvertPtr(argIn.Ptr, reflect.ValueOf(user)), nil
		}

		return reflect.Value{}, nil
	}))

	r := adapt.Adapt(gin.New(), af)

	doTest(t, r, af)
}

func doTest(t *testing.T, r *adapt.Adaptee, af *anyfn.Adapter) {
	r.GET("/GetAge1/:name", af.F(func(user AuthUser, name string) string {
		return user.Name + "/" + name
	}))
	r.GET("/GetAge2/:name", af.F(func(name string, user AuthUser) string {
		return user.Name + "/" + name
	}))
	r.GET("/GetAge3/:name", af.F(func(user *AuthUser, name string) string {
		return user.Name + "/" + name
	}))
	r.GET("/GetAge4/:name", af.F(func(name string, user *AuthUser) string {
		return user.Name + "/" + name
	}))

	// r.Run(":8080")

	rr := gintest.Get("/GetAge1/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
	rr = gintest.Get("/GetAge2/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
	rr = gintest.Get("/GetAge3/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
	rr = gintest.Get("/GetAge4/bingoohuang", r)
	assert.Equal(t, "TestAuthUser/bingoohuang", rr.Body())
}
