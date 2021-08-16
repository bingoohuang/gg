package anyfn

import (
	"github.com/bingoohuang/gg/pkg/ginx"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

func (a *Adapter) PrependInSupport(support InSupport) {
	a.InSupports = append([]InSupport{support}, a.InSupports...)
}

func (a *Adapter) PrependOutSupport(support OutSupport) {
	a.OutSupports = append([]OutSupport{support}, a.OutSupports...)
}

type InSupport interface {
	InSupport(argIn ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error)
}

type InSupportFn func(argIn ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error)

func (f InSupportFn) InSupport(argIn ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	return f(argIn, argsIn, c)
}

var GinContextType = reflect.TypeOf((*gin.Context)(nil)).Elem()
var InvalidValue = reflect.Value{}

func GinContextSupport(arg ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	if arg.Ptr && arg.Type == GinContextType { // 直接注入gin.Context
		return reflect.ValueOf(c), nil
	}

	return InvalidValue, nil
}

var HTTPRequestType = reflect.TypeOf((*http.Request)(nil)).Elem()

func HTTPRequestSupport(arg ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	if arg.Ptr && arg.Type == HTTPRequestType {
		return reflect.ValueOf(c.Request), nil
	}

	return InvalidValue, nil
}

var HTTPResponseWriterType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()

func HTTPResponseWriterSupport(arg ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	if arg.Kind == reflect.Interface && arg.Type == HTTPResponseWriterType {
		return reflect.ValueOf(c.Writer), nil
	}

	return InvalidValue, nil
}

func ContextKeyValuesSupport(arg ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	for _, v := range c.Keys {
		if arg.Type == NonPtrTypeOf(v) {
			return ConvertPtr(arg.Ptr, reflect.ValueOf(v)), nil
		}
	}

	return InvalidValue, nil
}

func BindSupport(arg ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	argValue := reflect.New(arg.Type)
	if err := ginx.ShouldBind(c, argValue.Interface()); err != nil {
		return InvalidValue, &AdapterError{Err: err, Context: "ShouldBind"}
	}

	return ConvertPtr(arg.Ptr, argValue), nil
}

func SinglePrimitiveValueSupport(arg ArgIn, argsIn []ArgIn, c *gin.Context) (reflect.Value, error) {
	v := singlePrimitiveValue(c, argsIn)
	if v != "" {
		return arg.convertValue(v)
	}

	return InvalidValue, nil
}
