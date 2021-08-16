package hlog

import (
	"bytes"
	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"time"

	"github.com/bingoohuang/gg/pkg/snow"
	"github.com/gin-gonic/gin"
)

type Adapter struct {
	Store Store
}

func NewAdapter(store Store) *Adapter {
	adapter := &Adapter{
		Store: store,
	}

	return adapter
}

type Middle struct {
	hlog         *hlog
	relativePath string
	P            *Adapter
}

func (m *Middle) Handle(c *gin.Context) {}

type writer struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *writer) Write(data []byte) (n int, err error) {
	w.buf.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *writer) WriteString(s string) (n int, err error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *writer) Body(maxSize int) string {
	if w.ResponseWriter.Size() <= maxSize {
		return w.buf.String()
	}

	return string(w.buf.Bytes()[:maxSize-3]) + "..."
}

func (m *Middle) Before(c *gin.Context) (after adapt.Handler) {
	if m.P.Store == nil || m.hlog.Option.Ignore {
		return nil
	}

	l := &Log{Created: time.Now()}

	l.Option = m.hlog.Option
	l.PathParams = c.Params
	l.Biz = l.Option.Biz

	r := c.Request
	l.Method = r.Method
	l.URL = r.URL.String()
	l.ReqHeader = r.Header
	l.Request = r

	l.ID = snow.Next().String()
	l.IPAddr = GetRemoteAddress(r)

	maxSize := m.hlog.Option.MaxSize
	l.ReqBody = string(PeekBody(r, maxSize))

	copyWriter := &writer{
		ResponseWriter: c.Writer,
	}
	c.Writer = copyWriter
	l.Start = time.Now()

	return adapt.HandlerFunc(func(c *gin.Context) {
		l.End = time.Now()
		l.Duration = l.End.Sub(l.Start)
		l.RspStatus = copyWriter.Status()
		l.RespSize = copyWriter.Size()
		l.RspBody = copyWriter.Body(maxSize)
		l.RspHeader = copyWriter.Header()
		l.Attrs = c.Keys

		m.P.Store.Store(c, l)
	})
}

func (a *Adapter) Default(relativePath string) adapt.Handler {
	return &Middle{
		relativePath: relativePath,
		hlog: &hlog{
			Option: NewOption(),
		},
		P: a,
	}
}

func (a *Adapter) Adapt(relativePath string, argV interface{}) adapt.Handler {
	hlog, ok := argV.(*hlog)
	if !ok {
		return nil
	}

	middle := &Middle{
		relativePath: relativePath,
		hlog:         hlog,
		P:            a,
	}

	return middle
}

type Option struct {
	MaxSize int
	Biz     string
	Ignore  bool
	Tables  []string
}

func NewOption() *Option {
	return &Option{
		MaxSize: 3000,
	}
}

type OptionFn func(option *Option)

func (a *Adapter) MaxSize(v int) OptionFn {
	return func(option *Option) {
		option.MaxSize = v
	}
}

func (a *Adapter) Tables(tables ...string) OptionFn {
	return func(option *Option) {
		option.Tables = tables
	}
}

func (a *Adapter) Biz(biz string) OptionFn {
	return func(option *Option) {
		option.Biz = biz
	}
}

func (a *Adapter) Ignore() OptionFn {
	return func(option *Option) {
		option.Ignore = true
	}
}

type hlog struct {
	Option *Option
	P      *Adapter
}

func (h *hlog) Parent() adapt.Adapter { return h.P }

func (a *Adapter) F(fns ...OptionFn) adapt.Parent {
	o := NewOption()
	for _, f := range fns {
		f(o)
	}

	return &hlog{P: a, Option: o}
}
