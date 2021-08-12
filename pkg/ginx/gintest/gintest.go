package gintest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
)

func Get(target string, router http.Handler, fns ...VarsFn) *Response {
	return Request(http.MethodGet, target, router, fns...)
}

func Post(target string, router http.Handler, fns ...VarsFn) *Response {
	return Request(http.MethodPost, target, router, fns...)
}

// from https://github.com/gin-gonic/gin/issues/1120
func Request(method, target string, router http.Handler, fns ...VarsFn) *Response {
	vars := &Vars{
		Query: make(map[string]string),
	}

	for _, fn := range fns {
		fn(vars)
	}

	r := httptest.NewRequest(method, target, vars.Body)

	if len(vars.Query) > 0 {
		q := r.URL.Query()
		for k, v := range vars.Query {
			q.Add(k, v)
		}

		r.URL.RawQuery = q.Encode()
	}

	if vars.ContentType != "" {
		r.Header.Set("Content-Type", vars.ContentType)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return &Response{
		ResponseRecorder: w,
	}
}

type Response struct {
	*httptest.ResponseRecorder
	body []byte
}

func (r *Response) Body() string {
	if r.body == nil {
		r.body, _ = ioutil.ReadAll(r.ResponseRecorder.Body)
	}

	return string(r.body)
}

func (r *Response) StatusCode() int {
	return r.ResponseRecorder.Code
}

type Vars struct {
	Body        io.Reader
	ContentType string
	Query       map[string]string
}

type VarsFn func(r *Vars)

func Query(k, v string) VarsFn {
	return func(r *Vars) {
		r.Query[k] = v
	}
}

func JSONVar(s interface{}) VarsFn {
	switch v := s.(type) {
	case string:
		return func(r *Vars) {
			r.Body = strings.NewReader(v)
			r.ContentType = "application/json; charset=utf-8"
		}
	default:
		return func(r *Vars) {
			b, _ := json.Marshal(v)
			r.Body = bytes.NewReader(b)
			r.ContentType = "application/json; charset=utf-8"
		}
	}
}
