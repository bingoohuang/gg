package hlog_test

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type A interface {
	Do() string
}

type obj1 struct {
	age int
}

func (a *obj1) Do() string { return fmt.Sprintf("%d", a.age) }

type obj2 struct {
	name string
}

func (a *obj2) Do() string { return a.name }

type obj3 struct {
	A
	j int64 // nolint:structcheck,unused
}

func TestReplaceResponseWriter(t *testing.T) {
	var o1 A = &obj3{A: &obj1{age: 100}}
	assert.Equal(t, "100", o1.Do())

	no1 := unsafe.Pointer(&o1)
	no1a := (*A)(no1)
	*no1a = &obj2{name: "bingoo"}
	o1.Do()

	assert.Equal(t, "bingoo", o1.Do())

	type Num struct {
		i string
		j int64
	}

	n := Num{i: "EDDYCJY", j: 1}
	nPointer := unsafe.Pointer(&n)

	niPointer := (*string)(nPointer)
	*niPointer = "煎鱼"

	njPointer := (*int64)(unsafe.Pointer(uintptr(nPointer) + unsafe.Offsetof(n.j)))
	*njPointer = 2

	assert.Equal(t, "n.i: 煎鱼, n.j: 2", fmt.Sprintf("n.i: %s, n.j: %d", n.i, n.j))
}
