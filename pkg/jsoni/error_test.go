package jsoni

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"unsafe"

	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
)

type JSONError struct {
	Error error
}

func (e JSONError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error.Error())
}

type errorEncoder struct{}

func (e errorEncoder) IsEmpty(ctx context.Context, ptr unsafe.Pointer, checkZero bool) bool {
	err := *((*error)(ptr))
	return err == nil
}

func (e errorEncoder) Encode(ctx context.Context, ptr unsafe.Pointer, stream *Stream) {
	s := *((*error)(ptr))
	q := strconv.Quote(s.Error())
	stream.WriteRaw(q)
}

type A struct {
	Err JSONError `json:"err"`
}

type B struct {
	Err error `json:"err"`
}

func TestError(t *testing.T) {
	j, _ := json.Marshal(A{Err: JSONError{errors.New("err")}})
	assert.Equal(t, `{"err":"err"}`, string(j))

	jc := Config{
		EscapeHTML: true,
	}.Froze()
	jc.RegisterTypeEncoder(reflect2.TypeOfPtr((*error)(nil)).Elem().String(), &errorEncoder{})

	j, _ = jc.Marshal(context.TODO(), B{Err: errors.New("err")})
	assert.Equal(t, `{"err":"err"}`, string(j))
}
