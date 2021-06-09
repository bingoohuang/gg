package badgerdb

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	db, err := New("abc", true)
	assert.Nil(t, err)

	for i := uint64(0); i < 100; i++ {
		err = db.Set(Uint64ToBytes(i), []byte(fmt.Sprintf("hello:%d", i)))
		assert.Nil(t, err)
	}

	v, err := db.Get(Uint64ToBytes(99))
	assert.Nil(t, err)
	assert.Equal(t, "hello:99", string(v))

	err = db.Walk(func(k, v []byte) error {
		fmt.Printf("%d=%s\n", BytesToUint64(k), v)
		return nil
	})
	assert.Nil(t, err)
	assert.Nil(t, db.Close())
}
