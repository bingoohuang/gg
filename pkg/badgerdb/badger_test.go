package badgerdb

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/bytex"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	db, err := Open(WithInMemory(true))
	assert.Nil(t, err)

	for i := uint64(0); i < 100; i++ {
		err = db.Set(bytex.FromUint64(i), []byte(fmt.Sprintf("hello:%d", i)))
		assert.Nil(t, err)
	}

	v, err := db.Get(bytex.FromUint64(99))
	assert.Nil(t, err)
	assert.Equal(t, "hello:99", string(v))

	err = db.Walk(func(k, v []byte) error {
		fmt.Printf("%d=%s\n", bytex.ToUint64(k), v)
		return nil
	})
	assert.Nil(t, err)
	assert.Nil(t, db.Close())
}
