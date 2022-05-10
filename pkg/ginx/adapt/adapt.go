package adapt

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Adapter interface {
	Adapt(relativePath string, arg interface{}) Handler
	Default(relativePath string) Handler
}

type Parent interface {
	Parent() Adapter
}

type Adaptee struct {
	Router       Gin
	adapterFuncs map[reflect.Type]*adapterFuncItem
	adapters     []Adapter
}

type adapterFuncItem struct {
	adapterFunc reflect.Value
}

type Gin interface {
	gin.IRouter
	http.Handler
}

// Handler defines the handler used by gin middleware as return value.
type Handler interface {
	Handle(*gin.Context)
}

type Handlers []Handler

func (h Handlers) Handle(c *gin.Context) {
	for _, e := range h {
		if e != nil {
			e.Handle(c)
		}
	}
}

type HandlerFunc gin.HandlerFunc

func (h HandlerFunc) Handle(c *gin.Context) { h(c) }

type Middleware interface {
	Before(c *gin.Context) (after Handler)
}

type Middlewares []Middleware

func (s Middlewares) Before(c *gin.Context) (after Handler) {
	afters := make([]Handler, len(s))
	for i, m := range s {
		afters[i] = m.Before(c)
	}

	return Handlers(afters)
}

func (i *adapterFuncItem) invoke(adapteeFn interface{}) gin.HandlerFunc {
	args := []reflect.Value{reflect.ValueOf(adapteeFn)}
	result := i.adapterFunc.Call(args)
	return result[0].Convert(GinHandlerFuncType).Interface().(gin.HandlerFunc)
}

func (a *Adaptee) createHandlerFuncs(relativePath string, args []interface{}) gin.HandlerFunc {
	adapterUnused := make(map[Adapter]bool)
	for _, ad := range a.adapters {
		adapterUnused[ad] = true
	}

	hfs := make([]Handler, 0, len(args))
	middlewares := make([]Middleware, 0, len(a.adapters))

	for _, arg := range args {
		hf, ms := a.adapt(relativePath, arg, adapterUnused)
		if hf != nil {
			hfs = append(hfs, hf)
		}

		if len(ms) > 0 {
			middlewares = append(middlewares, ms...)
		}
	}

	for k := range adapterUnused {
		f := k.Default(relativePath)
		if f == nil {
			continue
		}

		if m, ok := f.(Middleware); ok {
			middlewares = append(middlewares, m)
		}

		hfs = append(hfs, f)
	}

	return func(c *gin.Context) {
		defer Middlewares(middlewares).Before(c).Handle(c)

		Handlers(hfs).Handle(c)
	}
}

func (a *Adaptee) adapt(relativePath string, arg interface{}, adapterUnused map[Adapter]bool) (Handler, Middlewares) {
	if arg == nil {
		return nil, nil
	}

	if v := reflect.ValueOf(arg); v.Type().ConvertibleTo(GinHandlerFuncType) {
		return HandlerFunc(v.Convert(GinHandlerFuncType).Interface().(gin.HandlerFunc)), nil
	}

	if f := a.findAdapterFunc(arg); f != nil {
		return f, nil
	}

	return a.findAdapter(relativePath, arg, adapterUnused)
}

func (a *Adaptee) findAdapterFunc(arg interface{}) HandlerFunc {
	argType := reflect.TypeOf(arg)

	for funcType, funcItem := range a.adapterFuncs {
		if argType.ConvertibleTo(funcType) {
			return HandlerFunc(funcItem.invoke(arg))
		}
	}

	return nil
}

func (a *Adaptee) findAdapter(relativePath string, arg interface{}, adapterUnused map[Adapter]bool) (HandlerFunc, Middlewares) {
	chain := make([]Handler, 0, len(a.adapters))
	middlewares := make([]Middleware, 0, len(a.adapters))

	for _, v := range a.adapters {
		if f := v.Adapt(relativePath, arg); f != nil {
			delete(adapterUnused, v)

			if m, ok := f.(Middleware); ok {
				middlewares = append(middlewares, m)
			}

			chain = append(chain, f)
		}
	}

	return Handlers(chain).Handle, middlewares
}

func Adapt(router *gin.Engine, adapters ...interface{}) *Adaptee {
	a := &Adaptee{
		Router:       router,
		adapterFuncs: make(map[reflect.Type]*adapterFuncItem),
	}

	for _, adapt := range adapters {
		a.RegisterAdapter(adapt)
	}

	return a
}

func (a *Adaptee) ServeHTTP(r http.ResponseWriter, w *http.Request) {
	a.Router.ServeHTTP(r, w)
}

var GinHandlerFuncType = reflect.TypeOf(gin.HandlerFunc(nil))

func (a *Adaptee) RegisterAdapter(adapter interface{}) {
	if v, ok := adapter.(Adapter); ok {
		a.adapters = append(a.adapters, v)
		return
	}

	adapterValue := reflect.ValueOf(adapter)
	t := adapterValue.Type()

	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("register method should use a func"))
	}

	if t.NumIn() != 1 || t.In(0).Kind() != reflect.Func {
		panic(fmt.Errorf("register method should use a func which inputs a func"))
	}

	if t.NumOut() != 1 || !t.Out(0).ConvertibleTo(GinHandlerFuncType) {
		panic(fmt.Errorf("register method should use a func which returns gin.HandlerFunc"))
	}

	a.adapterFuncs[t.In(0)] = &adapterFuncItem{
		adapterFunc: adapterValue,
	}
}

func (a *Adaptee) Use(f func(c *gin.Context)) {
	a.Router.Use(f)
}

func (a *Adaptee) Handle(httpMethod, relativePath string, args ...interface{}) {
	a.Router.Handle(httpMethod, relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) Any(relativePath string, args ...interface{}) {
	a.Router.Any(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) POST(relativePath string, args ...interface{}) {
	a.Router.POST(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) GET(relativePath string, args ...interface{}) {
	a.Router.GET(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) DELETE(relativePath string, args ...interface{}) {
	a.Router.DELETE(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) PUT(relativePath string, args ...interface{}) {
	a.Router.PUT(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) PATCH(relativePath string, args ...interface{}) {
	a.Router.PATCH(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) OPTIONS(relativePath string, args ...interface{}) {
	a.Router.OPTIONS(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) HEAD(relativePath string, args ...interface{}) {
	a.Router.HEAD(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *Adaptee) Group(relativePath string, args ...interface{}) *AdapteeGroup {
	g := a.Router.Group(relativePath, a.createHandlerFuncs(relativePath, args))
	return &AdapteeGroup{
		Adaptee:     a,
		RouterGroup: g,
	}
}

type AdapteeGroup struct {
	*Adaptee
	*gin.RouterGroup
}

func (a *AdapteeGroup) Use(f func(c *gin.Context)) {
	a.RouterGroup.Use(f)
}

func (a *AdapteeGroup) Any(relativePath string, args ...interface{}) {
	a.RouterGroup.Any(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) POST(relativePath string, args ...interface{}) {
	a.RouterGroup.POST(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) GET(relativePath string, args ...interface{}) {
	a.RouterGroup.GET(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) DELETE(relativePath string, args ...interface{}) {
	a.Router.DELETE(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) PUT(relativePath string, args ...interface{}) {
	a.RouterGroup.PUT(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) PATCH(relativePath string, args ...interface{}) {
	a.RouterGroup.PATCH(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) OPTIONS(relativePath string, args ...interface{}) {
	a.RouterGroup.OPTIONS(relativePath, a.createHandlerFuncs(relativePath, args))
}

func (a *AdapteeGroup) HEAD(relativePath string, args ...interface{}) {
	a.RouterGroup.HEAD(relativePath, a.createHandlerFuncs(relativePath, args))
}
