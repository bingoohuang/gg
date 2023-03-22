package ginx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/bingoohuang/gg/pkg/jsoni/extra"
	"github.com/bingoohuang/gg/pkg/strcase"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ShouldBind checks the Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
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
	JsoniConfig.RegisterExtension(&extra.NamingStrategyExtension{Translate: strcase.ToCamelLower})
}

func decodeJSON(r io.Reader, obj interface{}) error {
	if err := JsoniConfig.NewDecoder(r).Decode(nil, obj); err != nil {
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
	Data     interface{}
	JsoniAPI jsoni.API
}

var jsonContentType = []string{"application/json; charset=utf-8"}

// Render (JSON) writes data with custom ContentType.
func (r JSONRender) Render(w http.ResponseWriter) (err error) {
	return WriteJSONOptions(w, r.Data, WithJsoniAPI(r.JsoniAPI))
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSONRender) WriteContentType(w http.ResponseWriter) { writeContentType(w, jsonContentType) }

type WriteJSONConfig struct {
	JsoniAPI jsoni.API
}

type WriteJSONConfigFn func(*WriteJSONConfig)

func WithJsoniAPI(api jsoni.API) WriteJSONConfigFn {
	return func(o *WriteJSONConfig) {
		o.JsoniAPI = api
	}
}

// WriteJSONOptions marshals the given interface object and writes it with custom ContentType.
func WriteJSONOptions(w http.ResponseWriter, obj interface{}, fns ...WriteJSONConfigFn) error {
	options := &WriteJSONConfig{}
	for _, fn := range fns {
		fn(options)
	}

	if options.JsoniAPI == nil {
		options.JsoniAPI = JsoniConfig
	}

	writeContentType(w, jsonContentType)

	data, err := options.JsoniAPI.Marshal(nil, obj)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := JsoniConfig.Marshal(nil, obj)
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
