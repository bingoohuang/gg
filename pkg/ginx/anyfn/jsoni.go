package anyfn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"net/http"
)

// ShouldBind checks the Content-Type to select a binding engine automatically,
// Depending the "Content-Type" header different bindings are used:
//     "application/json" --> JSON binding
//     "application/xml"  --> XML binding
// otherwise --> returns an error
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// Like c.Bind() but this method does not set the response status code to 400 and abort if the json is not valid.
func ShouldBind(c *gin.Context, obj interface{}) error {
	var b binding.Binding
	switch c.ContentType() {
	case binding.MIMEJSON:
		b = JSONBind
	default:
		b = binding.Default(c.Request.Method, c.ContentType())
	}
	return c.ShouldBindWith(obj, b)
}

var JSONBind = jsoniBinding{}

type jsoniBinding struct{}

func (jsoniBinding) Name() string { return "json" }

func (jsoniBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return decodeJSON(req.Body, obj)
}

func (jsoniBinding) BindBody(body []byte, obj interface{}) error {
	return decodeJSON(bytes.NewReader(body), obj)
}

// JsoniConfig tries to be 100% compatible with standard library behavior
var JsoniConfig = jsoni.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	Int64AsString:          true,
}.Froze()

func init() {
	JsoniConfig.RegisterExtension(&extra.NamingStrategyExtension{Translate: extra.LowerCaseWithUnderscores})
}

func decodeJSON(r io.Reader, obj interface{}) error {
	decoder := JsoniConfig.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}

func validate(obj interface{}) error {
	validator := binding.Validator
	if validator == nil {
		return nil
	}
	return validator.ValidateStruct(obj)
}

// JSONRender contains the given interface object.
type JSONRender struct {
	Data interface{}
}

var jsonContentType = []string{"application/json; charset=utf-8"}

// Render (JSON) writes data with custom ContentType.
func (r JSONRender) Render(w http.ResponseWriter) (err error) {
	if err = WriteJSON(w, r.Data); err != nil {
		panic(err)
	}
	return
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSONRender) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
