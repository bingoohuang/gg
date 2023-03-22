package sqx_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/sqlparse/sqlparser"
	"github.com/bingoohuang/gg/pkg/sqx"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestQueryAsBeans(t *testing.T) {
	// 创建数据库连接池
	_, db, err := sqx.Open("sqlite3", ":memory:")
	assert.Nil(t, err)

	effectedRows, err := sqx.NewSQL("create table person(id varchar(100), age int)").Update(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), effectedRows)
	effectedRows, err = sqx.NewSQL("insert into person(id, age) values(?)", "嫦娥", 1000).Update(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), effectedRows)
	effectedRows, err = sqx.NewSQL("insert into person(id, age) values(?, ?)", "悟空", 500).Update(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), effectedRows)

	m, err := sqx.NewSQL("select id, age from person where id=?", "嫦娥").QueryAsMap(db)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"age": "1000", "id": "嫦娥"}, m)

	r, err := sqx.NewSQL("select id, age from person where id=?", "嫦娥").QueryAsRow(db)
	assert.Nil(t, err)
	assert.Equal(t, []string{"嫦娥", "1000"}, r)

	var ids []string
	err = sqx.NewSQL("select id from person order by age").Query(db, &ids)
	assert.Nil(t, err)
	assert.Equal(t, []string{"悟空", "嫦娥"}, ids)

	var ages []int
	err = sqx.NewSQL("select age from person order by age").Query(db, &ages)
	assert.Nil(t, err)
	assert.Equal(t, []int{500, 1000}, ages)

	var age1 int
	err = sqx.NewSQL("select age from person order by age").Query(db, &age1)
	assert.Nil(t, err)
	assert.Equal(t, 500, age1)

	type Person struct {
		ID string
		Ag int `col:"AGE"`
	}

	var ps []Person
	err = sqx.NewSQL("select id, age from person where id=?", "嫦娥").Query(db, &ps)
	assert.Nil(t, err)
	assert.Equal(t, []Person{{ID: "嫦娥", Ag: 1000}}, ps)

	err = sqx.NewSQL("select id, age from person where id in(?) order by age desc", "嫦娥", "悟空").Query(db, &ps)
	assert.Nil(t, err)
	assert.Equal(t, []Person{{ID: "嫦娥", Ag: 1000}, {ID: "悟空", Ag: 500}}, ps)

	ps = nil
	err = sqx.NewSQL("select id, age from person where id in(?) order by age desc").Query(db, &ps)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	assert.Equal(t, 0, len(ps))

	var p Person
	err = sqx.NewSQL("select id, age from person where id=?", "嫦娥").Query(db, &p)
	assert.Nil(t, err)
	assert.Equal(t, Person{ID: "嫦娥", Ag: 1000}, p)

	var p2 *Person
	err = sqx.NewSQL("select id, age from person where id=?", "嫦娥").Query(db, &p2)
	assert.Nil(t, err)
	assert.Equal(t, Person{ID: "嫦娥", Ag: 1000}, *p2)

	var p3 *Person
	err = sqx.NewSQL("select id, age from person where id=?", "八戒").Query(db, &p3)
	assert.Nil(t, p3)
	assert.True(t, errors.Is(err, sql.ErrNoRows))

	age, err := sqx.NewSQL("select age from person where id=?", "嫦娥").QueryAsNumber(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), age)

	id, err := sqx.NewSQL("select id from person where id=?", "嫦娥").QueryAsString(db)
	assert.Nil(t, err)
	assert.Equal(t, "嫦娥", id)
}

func TestDriverName(t *testing.T) {
	pg := "postgres://SYSTEM:111111@192.168.126.245:54322/METRICS_UMP?sslmode=disable"
	// 创建数据库连接池
	db, d, err := sqx.Open("pgx", pg)
	assert.Nil(t, err)
	assert.Equal(t, "pgx", sqx.DriverName(db.Driver()))
	assert.Equal(t, sqlparser.Postgresql, d.GetDBType())

	sq := sqx.SQL{
		Q:    `INSERT INTO COMPANY (NAME,AGE,ADDRESS,SALARY) VALUES(:?)`,
		Vars: []interface{}{"Paul", 32, "California", 20000.00},
	}
	if d.GetDBType() == sqlparser.Postgresql || d.GetDBType() == sqlparser.Kingbase {
		sq.AppendQ = `RETURNING ID`
	}

	id, err := sq.QueryAsNumber(d)
	fmt.Println(id, err)
}

func TestQuery(t *testing.T) {
	raw, db := openDB(t)
	assert.Equal(t, "sqlite3", sqx.DriverName(raw.Driver()))

	_, err := sqx.SQL{Q: "create table person(id varchar(100), age int)"}.Update(db)
	assert.Nil(t, err)
	s := sqx.SQL{Q: "insert into person(id, age) values(?)"}
	_, err = s.WithVars("嫦娥", 1000).Update(db)
	assert.Nil(t, err)
	_, err = s.WithVars("悟空", 500).Update(db)
	assert.Nil(t, err)

	s = sqx.SQL{Q: "select id, age from person"}
	m, err := s.Append("where id=?", "嫦娥").QueryAsMap(db)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"id": "嫦娥", "age": "1000"}, m)

	type Person struct {
		ID string
		Ag int `col:"AGE"`
	}

	var ps []Person
	assert.Nil(t, s.Query(db, &ps))
	assert.Equal(t, []Person{{ID: "嫦娥", Ag: 1000}}, ps)

	var p Person
	assert.Nil(t, s.Query(db, &p))
	assert.Equal(t, Person{ID: "嫦娥", Ag: 1000}, p)

	s = sqx.SQL{Q: "select age from person"}
	age, err := s.Append("where id=?", "嫦娥").QueryAsNumber(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), age)

	var ageValue int

	err = s.QueryRaw(db, sqx.WithScanRow(func(columns []string, rows *sql.Rows, _ int) (bool, error) {
		return false, rows.Scan(&ageValue)
	}))
	assert.Nil(t, err)
	assert.Equal(t, 1000, ageValue)
}
