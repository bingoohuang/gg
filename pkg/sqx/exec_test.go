package sqx_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/sqx"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// nolint
func ExampleExec() {
	// 创建数据库连接池
	db, _ := sql.Open("sqlite3", ":memory:")

	sqx.NewSQL("create table person(id varchar(100), age int)").Update(db)
	sqx.NewSQL("insert into person(id, age) values(?, ?)", "嫦娥", 1000).Update(db)
	sqx.NewSQL("insert into person(id, age) values(?, ?)", "悟空", 500).Update(db)

	m, _ := sqx.NewSQL("select id, age from person where id=?", "嫦娥").QueryAsMap(db)
	fmt.Println(m) // // map[age:1000 id:嫦娥]

	r, _ := sqx.NewSQL("select id, age from person where id=?", "嫦娥").QueryAsRow(db)
	fmt.Println(r) // // [嫦娥 1000]

	type Person struct {
		ID string
		Ag int `col:"AGE"`
	}

	var ps []Person
	sqx.NewSQL("select id, age from person where id=?", "嫦娥").QueryAsBeans(db, &ps)
	fmt.Printf("%+v\n", ps) // [{ID:嫦娥 Ag:1000}]

	var p Person
	sqx.NewSQL("select id, age from person where id=?", "嫦娥").QueryAsBeans(db, &p)
	fmt.Printf("%+v\n", p) // {ID:嫦娥 Ag:1000}

	age, _ := sqx.NewSQL("select age from person where id=?", "嫦娥").QueryAsNumber(db)
	fmt.Println(age) // 1000

	id, _ := sqx.NewSQL("select id from person where id=?", "嫦娥").QueryAsString(db)
	fmt.Println(id) // 嫦娥

	// Output:
	// map[age:1000 id:嫦娥]
	// [嫦娥 1000]
	// [{ID:嫦娥 Ag:1000}]
	// {ID:嫦娥 Ag:1000}
	// 1000
	// 嫦娥
}

func TestDriverName(t *testing.T) {
	pg := "postgres://SYSTEM:111111@192.168.126.245:54322/METRICS_UMP?sslmode=disable"
	// 创建数据库连接池
	db, err := sql.Open("pgx", pg)
	assert.Nil(t, err)
	assert.Equal(t, "pgx", sqx.DriverName(db))
}

func TestQuery(t *testing.T) {
	db := openDB(t)
	assert.Equal(t, "sqlite3", sqx.DriverName(db))

	_, err := sqx.SQL{Query: "create table person(id varchar(100), age int)"}.Update(db)
	assert.Nil(t, err)
	s := sqx.SQL{Query: "insert into person(id, age) values(?, ?)"}
	_, err = s.WithVars("嫦娥", 1000).Update(db)
	assert.Nil(t, err)
	_, err = s.WithVars("悟空", 500).Update(db)
	assert.Nil(t, err)

	s = sqx.SQL{Query: "select id, age from person"}
	m, err := s.Append("where id=?", "嫦娥").QueryAsMap(db)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"id": "嫦娥", "age": "1000"}, m)

	type Person struct {
		ID string
		Ag int `col:"AGE"`
	}

	var ps []Person
	assert.Nil(t, s.QueryAsBeans(db, &ps))
	assert.Equal(t, []Person{{ID: "嫦娥", Ag: 1000}}, ps)

	var p Person
	assert.Nil(t, s.QueryAsBeans(db, &p))
	assert.Equal(t, Person{ID: "嫦娥", Ag: 1000}, p)

	s = sqx.SQL{Query: "select age from person"}
	age, err := s.Append("where id=?", "嫦娥").QueryAsNumber(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), age)

	var ageValue int

	err = s.QueryRaw(db, sqx.WithScanRow(func(rows *sql.Rows, _ int) (bool, error) {
		return false, rows.Scan(&ageValue)
	}))
	assert.Nil(t, err)
	assert.Equal(t, 1000, ageValue)
}
