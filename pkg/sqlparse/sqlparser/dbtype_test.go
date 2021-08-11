package sqlparser

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func TestConvertDBType(t *testing.T) {
	const q0 = "select `name` from `user` where `min` >= ? and name like 'abc?' order by age"
	const q1 = `select "name" from "user" where "min" >= $1 and "name" like 'abc?' order by "age"`
	const q2 = `select count(*) as "cnt" from "user" where "min" >= $1 and "name" like 'abc?'`

	q, r, err := Kingbase.Convert(q0)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, q1, q)

	v := &Paging{PageSeq: 1, PageSize: 20, RowsCount: CreateCountingQuery}
	q, r, err = Kingbase.Convert(q0, WithPaging(v))
	assert.Nil(t, err)
	assert.Equal(t, q1+" limit $2 offset $3", q)
	assert.Equal(t, q2, r.CountingQuery)
	assert.Equal(t, []interface{}{20, 0}, r.ExtraArgs)

	const q3 = `insert into t (a, b, c, d, e, f,g) values(?, ?, ?, ?, ?, ?,?) on duplicate key update a=values(a), b=values(b), c=values(c),d=values(c)`

	q, r, err = Kingbase.Convert(q3)
	assert.True(t, errors.Is(err, ErrSyntax))

	// 在占位符数量与字段数量不匹配时，自动调整占位符数量
	const q40 = `insert into t (a, b, c) values(?,?)`
	const q41 = `insert into "t"("a", "b", "c") values ($1, $2, $3)`
	q, r, err = Kingbase.Convert(q40)
	assert.Nil(t, err)
	assert.Equal(t, q41, q)
	// 在占位符数量与字段数量不匹配时，自动调整占位符数量
	const q50 = `insert into t (a, b, c) values(?,?,?,?)`
	q, r, err = Kingbase.Convert(q50)
	assert.Nil(t, err)
	assert.Equal(t, ByPlaceholder, r.BindMode)
	assert.Equal(t, q41, q)

	const q60 = `insert into t (a, b, c) values(':a', ':b', ':c')`
	const q62 = `insert into t (a, b, c) values(:a, :b, :c)`
	const q61 = `insert into "t"("a", "b", "c") values ($1, $2, $3)`
	q, r, err = Kingbase.Convert(q60)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, q61, q)

	q, r, err = Kingbase.Convert(q62)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q61, q)

	const q70 = `insert into t (a, b, c) values(':2', ':1', ':3')` // b,a,c
	q, r, err = Kingbase.Convert(q70)
	assert.Nil(t, err)
	assert.Equal(t, []int{2, 1, 3}, r.VarPoses)
	assert.Equal(t, q61, q)

	const q71 = `insert into t (a, b, c) values(:1, :2, :3)`
	q, r, err = Kingbase.Convert(q71)
	assert.Nil(t, err)
	assert.Equal(t, []int{1, 2, 3}, r.VarPoses)
	assert.Equal(t, BySeq, r.BindMode)
	assert.Equal(t, q61, q)

	const q72 = `insert into t (a, b, c) values(:?, :?, :?)`
	q, r, err = Kingbase.Convert(q72)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q61, q)

	const q73 = `update t set a = :?, b = :? where c > :? and d = :?`
	const q74 = `update "t" set "a" = $1, "b" = $2 where "c" > $3 and "d" = $4`
	q, r, err = Kingbase.Convert(q73)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c", "d"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q74, q)
}
