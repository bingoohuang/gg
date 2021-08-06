package sqlparser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertDBType(t *testing.T) {
	const q0 = "select `name` from `user` where `min` >= ? and name like 'abc?'"
	const q1 = `select "name" from "user" where "min" >= $1 and "name" like 'abc?'`
	q, r, err := Kingbase.Convert(q0)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, q1, q)

	v := NewPaging()
	v.RowsCount = CreateCountingQuery
	q, r, err = Kingbase.Convert(q0, WithPaging(v))
	assert.Nil(t, err)
	assert.Equal(t, q1+" limit ? offset ?", q)
	assert.Equal(t, `select count(*) as "cnt" from "user" where "min" >= $2 and "name" like 'abc?'`, r.CountingQuery)
	assert.Equal(t, []interface{}{20, 0}, r.ExtraArgs)
}
