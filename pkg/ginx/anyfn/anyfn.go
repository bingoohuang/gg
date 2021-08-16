package anyfn

import (
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gg/pkg/ginx/adapt"

	"github.com/gin-gonic/gin"
)

// DirectDealer is the dealer for a specified type.
type DirectDealer interface {
	Deal(*gin.Context)
}

// HTTPStatus defines the type of HTTP state.
type HTTPStatus int

func (h HTTPStatus) Deal(c *gin.Context) {
	c.Status(int(h))
}

type Adapter struct {
	InSupports  []InSupport
	OutSupports []OutSupport
}

func (a *Adapter) Default(relativePath string) adapt.Handler { return nil }

func (a *Adapter) Adapt(relativePath string, argV interface{}) adapt.Handler {
	anyF, ok := argV.(*anyF)
	if !ok {
		return nil
	}

	fv := reflect.ValueOf(anyF.F)

	return adapt.HandlerFunc(func(c *gin.Context) {
		if err := a.internalAdapter(c, fv, anyF); err != nil {
			logrus.Warnf("adapt error %v", err)
		}
	})
}

func NewAdapter() *Adapter {
	adapter := &Adapter{
		InSupports: []InSupport{
			InSupportFn(GinContextSupport),
			InSupportFn(HTTPRequestSupport),
			InSupportFn(HTTPResponseWriterSupport),
			InSupportFn(ContextKeyValuesSupport),
			InSupportFn(SinglePrimitiveValueSupport), //  try single param
			InSupportFn(BindSupport),                 //  try bind
		},
		OutSupports: []OutSupport{
			OutSupportFn(ErrorSupport),
			OutSupportFn(DirectDealerSupport),
			OutSupportFn(DefaultSupport),
		},
	}

	return adapter
}

type Before interface {
	// Do will be called Before the adaptee invoking.
	Do(args []interface{}) error
}

type BeforeFn func(args []interface{}) error

func (b BeforeFn) Do(args []interface{}) error { return b(args) }

type After interface {
	// Do will be called Before the adaptee invoking.
	Do(args []interface{}, results []interface{}) error
}

type AfterFn func(args []interface{}, results []interface{}) error

func (b AfterFn) Do(args []interface{}, results []interface{}) error { return b(args, results) }

type anyF struct {
	P      *Adapter
	F      interface{}
	Option *Option
}

func (a *anyF) Parent() adapt.Adapter { return a.P }

type Option struct {
	Before Before
	After  After
	Attrs  map[string]interface{}
}

type OptionFn func(*Option)

func (a *Adapter) Before(before Before) OptionFn {
	return func(f *Option) {
		f.Before = before
	}
}

func (a *Adapter) After(after After) OptionFn {
	return func(f *Option) {
		f.After = after
	}
}

func (a *Adapter) AttrMap(attrs map[string]interface{}) OptionFn {
	return func(f *Option) {
		for k, v := range attrs {
			f.Attrs[k] = v
		}
	}
}

func (a *Adapter) Attr(k string, v interface{}) OptionFn {
	return func(f *Option) {
		f.Attrs[k] = v
	}
}

func (a *Adapter) F(v interface{}, fns ...OptionFn) *anyF {
	option := &Option{
		Attrs: make(map[string]interface{}),
	}

	for _, fn := range fns {
		fn(option)
	}

	return &anyF{F: v, Option: option}
}

func (a *Adapter) internalAdapter(c *gin.Context, fv reflect.Value, anyF *anyF) error {
	argVs, err := a.createArgs(c, fv)
	if err != nil {
		return err
	}

	if err := a.before(argVs, anyF.Option.Before); err != nil {
		return err
	}

	r := fv.Call(argVs)

	if err := a.after(argVs, r, anyF.Option.After); err != nil {
		return err
	}

	return a.processOut(c, fv, r)
}

func (a *Adapter) before(v []reflect.Value, f Before) error {
	if f == nil {
		return nil
	}

	return f.Do(Values(v).Interface())
}

func (a *Adapter) after(v, results []reflect.Value, f After) error {
	if f == nil {
		return nil
	}

	return f.Do(Values(v).Interface(), Values(results).Interface())
}

type Values []reflect.Value

func (v Values) Interface() []interface{} {
	args := make([]interface{}, len(v))
	for i, a := range v {
		args[i] = a.Interface()
	}

	return args
}

func (a *Adapter) processOut(c *gin.Context, fv reflect.Value, r []reflect.Value) error {
	numOut := fv.Type().NumOut()

	vs := make([]interface{}, 0, numOut)
	for i := 0; i < numOut; i++ {
		vs = append(vs, r[i].Interface())
	}

	if len(vs) > 0 {
		if err, ok := vs[len(vs)-1].(error); ok {
			a.processOutV(c, err, vs)
			return nil
		}
	}

	for _, v := range vs {
		a.processOutV(c, v, vs)
	}

	return nil
}

func (a *Adapter) processOutV(c *gin.Context, v interface{}, vs []interface{}) bool {
	for _, support := range a.OutSupports {
		if ok, _ := support.OutSupport(v, vs, c); ok {
			return ok
		}
	}

	return false
}

func (a *Adapter) createArgs(c *gin.Context, fv reflect.Value) (v []reflect.Value, err error) {
	ft := fv.Type()
	argIns := parseArgIns(ft)
	v = make([]reflect.Value, ft.NumIn())

	for i, arg := range argIns {
		if v[i], err = a.createArgValue(c, arg, argIns); err != nil {
			return nil, err
		}
	}

	return v, err
}

func (a *Adapter) createArgValue(c *gin.Context, arg ArgIn, argsIn []ArgIn) (reflect.Value, error) {
	for _, support := range a.InSupports {
		v, err := support.InSupport(arg, argsIn, c)
		if err == nil && v.IsValid() {
			return v, nil
		}

	}

	return reflect.Zero(arg.Type), nil
}
