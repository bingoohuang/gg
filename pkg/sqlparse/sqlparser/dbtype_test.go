package sqlparser

import (
	"errors"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestConvertDBType(t *testing.T) {
	const q0 = "select `name` from `user` where `min` >= ? and name like 'abc?' order by age"
	const q1 = `select "name" from "user" where "min" >= $1 and "name" like 'abc?' order by "age"`
	const q2 = `select count(*) as "cnt" from "user" where "min" >= $1 and "name" like 'abc?'`

	r, err := Kingbase.Convert(q0)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, q1, r.ConvertQuery())

	const q00 = `select * from user where name in (?, ?)`
	r, err = Kingbase.Convert(q00)
	assert.Nil(t, err)
	q, _ := r.PickArgs([]interface{}{"a", "b", "c"})
	assert.Equal(t, `select * from "user" where "name" in ($1, $2, $3)`, q)

	v := &Paging{PageSeq: 1, PageSize: 20, RowsCount: CreateCountingQuery}
	r, err = Kingbase.Convert(q0, WithPaging(v))
	assert.Nil(t, err)
	assert.Equal(t, q1+" limit $2 offset $3", r.ConvertQuery())
	assert.Equal(t, q2, r.CountingQuery)
	assert.Equal(t, []interface{}{20, 0}, r.ExtraArgs)

	const q3 = `insert into t (a, b, c, d, e, f,g) values(?, ?, ?, ?, ?, ?,?) on duplicate key update a=values(a), b=values(b), c=values(c),d=values(c)`

	r, err = Kingbase.Convert(q3)
	assert.True(t, errors.Is(err, ErrSyntax))

	// 在占位符数量与字段数量不匹配时，自动调整占位符数量
	const q40 = `insert into t (a, b, c) values(?,?)`
	const q41 = `insert into "t"("a", "b", "c") values ($1, $2, $3)`
	r, err = Kingbase.Convert(q40)
	assert.Nil(t, err)
	assert.Equal(t, q41, r.ConvertQuery())
	// 在占位符数量与字段数量不匹配时，自动调整占位符数量
	const q50 = `insert into t (a, b, c) values(?,?,?,?)`
	r, err = Kingbase.Convert(q50)
	assert.Nil(t, err)
	assert.Equal(t, ByPlaceholder, r.BindMode)
	assert.Equal(t, q41, r.ConvertQuery())

	const q60 = `insert into t (a, b, c) values(':a', ':b', ':c')`
	const q62 = `insert into t (a, b, c) values(:a, :b, :c)`
	const q63 = `insert into t (a, b, c) values(:?)`
	const q61 = `insert into "t"("a", "b", "c") values ($1, $2, $3)`
	r, err = Kingbase.Convert(q60)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, q61, r.ConvertQuery())

	r, err = Kingbase.Convert(q62)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q61, r.ConvertQuery())

	r, err = Kingbase.Convert(q63)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q61, r.ConvertQuery())

	const q70 = `insert into t (a, b, c) values(':2', ':1', ':3')` // b,a,c
	r, err = Kingbase.Convert(q70)
	assert.Nil(t, err)
	assert.Equal(t, []int{2, 1, 3}, r.VarPoses)
	assert.Equal(t, q61, r.ConvertQuery())

	const q71 = `insert into t (a, b, c) values(:1, :2, :3)`
	r, err = Kingbase.Convert(q71)
	assert.Nil(t, err)
	assert.Equal(t, []int{1, 2, 3}, r.VarPoses)
	assert.Equal(t, BySeq, r.BindMode)
	assert.Equal(t, q61, r.ConvertQuery())

	const q72 = `insert into t (a, b, c) values(:?, :?, :?)`
	r, err = Kingbase.Convert(q72)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q61, r.ConvertQuery())

	const q73 = `update t set a = :?, b = :? where c > :? and d = :?`
	const q74 = `update "t" set "a" = $1, "b" = $2 where "c" > $3 and "d" = $4`
	r, err = Kingbase.Convert(q73)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c", "d"}, r.VarNames)
	assert.Equal(t, ByName, r.BindMode)
	assert.Equal(t, q74, r.ConvertQuery())

	const q75 = `select id, item_code, item_name, group_code, group_name, status, sort, comment, create_time, update_time from sys_dictionary where group_code = 'monitor_kv_warn.metrics_type' order by sort`
	const q76 = `select id, item_code, item_name, group_code, group_name, status, "SORT", "COMMENT", create_time, update_time from sys_dictionary where group_code = 'monitor_kv_warn.metrics_type' order by "SORT"`
	r, err = Oracle.Convert(q75)
	assert.Nil(t, err)
	assert.Equal(t, q76, r.ConvertQuery())

	q77 := "select  id,name, rtx, `desc`, warn_level, create_time, update_time, account, type, initialized from warn_template order by update_time desc limit 10"
	q78 := `select id, name, rtx, "DESC", warn_level, create_time, update_time, account, "TYPE", initialized from warn_template order by update_time desc fetch next 10 rows only`
	r, err = Oracle.Convert(q77)
	assert.Nil(t, err)
	assert.Equal(t, q78, r.ConvertQuery())

	q79 := `select hi.server_ip as ip, mhi.name, mhi.version, mhi.status, mhi.timestamp from host_info hi left join (select m1.id, m1.ip, m1.name, m1.version, m1.status, timestamp from meta_heartbeat_info m1 join (select ip, name, max(timestamp) as max_time from meta_heartbeat_info group by ip, name) m2 on m1.ip = m2.ip and m1.name = m2.name and m1.timestamp = m2.max_time) mhi on hi.server_ip = mhi.ip`
	q80 := `select hi.server_ip as ip, mhi.name, mhi.version, mhi.status, mhi."TIMESTAMP" from host_info hi left join (select m1.id, m1.ip, m1.name, m1.version, m1.status, "TIMESTAMP" from meta_heartbeat_info m1 join (select ip, name, max("TIMESTAMP") as max_time from meta_heartbeat_info group by ip, name) m2 on m1.ip = m2.ip and m1.name = m2.name and m1."TIMESTAMP" = m2.max_time) mhi on hi.server_ip = mhi.ip`
	r, err = Oracle.Convert(q79)
	assert.Nil(t, err)
	assert.Equal(t, q80, r.ConvertQuery())

	q81 := `insert into meta_heartbeat_info(id, ip, name, version, status, TIMESTAMP) values ('192.168.127.109:meta-agent:1.0.0', '192.168.127.109', 'meta-agent', '1.0.0', 'started', '1724205143073')`
	q82 := `insert into meta_heartbeat_info(id, ip, name, version, status, "TIMESTAMP") values ('192.168.127.109:meta-agent:1.0.0', '192.168.127.109', 'meta-agent', '1.0.0', 'started', '1724205143073')`
	r, err = Oracle.Convert(q81)
	assert.Nil(t, err)
	assert.Equal(t, q82, r.ConvertQuery())
}
