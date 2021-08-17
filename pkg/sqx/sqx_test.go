package sqx_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/sqx"
	"github.com/stretchr/testify/assert"
)

func ExampleCondition() {
	type Cond struct {
		Name string `cond:"name like ?" modifier:"%v%"` // like的示例
		Addr string `cond:"addr = ?"`
		Code int    `cond:"code > ?" zero:"-1"` // Code == -1时，忽略本字段的条件
		Nooo string `cond:"-"`                  // 忽略本字段作为条件
	}

	ret, err := sqx.CreateSQL(`select code, name, addr, email from person order by code`, Cond{
		Name: "天问一号",
		Addr: "火星基地",
		Code: -1,
	})
	fmt.Println(fmt.Sprintf("%+v", ret), err)

	ret, err = sqx.CreateSQL(`select code, name, addr, email from person order by code`, Cond{
		Name: "嫦娥",
		Addr: "广寒宫",
		Code: 100,
	})
	fmt.Println(fmt.Sprintf("%+v", ret), err)

	ret.Append("limit ?, ?", 1, 10)
	fmt.Println(fmt.Sprintf("%+v", ret), err)

	// Output:
	// &{Q:select code, name, addr, email from person where name like ? and addr = ? order by code Vars:[%天问一号% 火星基地] Ctx:<nil> Log:false} <nil>
	// &{Q:select code, name, addr, email from person where name like ? and addr = ? and code > ? order by code Vars:[%嫦娥% 广寒宫 100] Ctx:<nil> Log:false} <nil>
	// &{Q:select code, name, addr, email from person where name like ? and addr = ? and code > ? order by code limit ?, ? Vars:[%嫦娥% 广寒宫 100 1 10] Ctx:<nil> Log:false} <nil>
}

func TestEmbeddedCondition(t *testing.T) {
	type Cond1 struct {
		BigName string
		C       int
		D       int
	}

	type Cond2 struct {
		Cond1
		E string
	}

	x, err := sqx.CreateSQL(`select a,b,c from t`, Cond2{Cond1: Cond1{BigName: "bb"}, E: "ee"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where big_name = ? and e = ?`, x.Q)
	assert.Equal(t, []interface{}{"bb", "ee"}, x.Vars)
}

func TestModifier(t *testing.T) {
	type Cond2 struct {
		E string `cond:"e like ?" modifier:"%v%"`
	}

	x, err := sqx.CreateSQL(`select a,b,c from t`, Cond2{E: "ee"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where e like ?`, x.Q)
	assert.Equal(t, []interface{}{"%ee%"}, x.Vars)
}

func TestManual(t *testing.T) {
	s := sqx.SQL{Q: `select * from warn_template_rule`}
	s.And(`source_type = ?`, "x")
	s.And(`temp_source_id = ?`, "y")
	s.And(`id in (?)`, []string{"a", "b", "c"})
	s.Append(`order by id desc`)
	assert.Equal(t, `select * from warn_template_rule where source_type = ? and temp_source_id = ? and id in (?,?,?) order by id desc`, s.Q)
	assert.Equal(t, []interface{}{"x", "y", "a", "b", "c"}, s.Vars)
}

func TestCondition(t *testing.T) {
	type cond struct {
		B    string // will generate `b = ?` when B is not zero.
		C    int    `cond:"c0 > ?"`           // use cond tag directly.
		D    int    `cond:"c2 > ?" zero:"-1"` // use cond tag directly when B is not specified zero.
		E    string `zero:"null"`
		Mark string `cond:"-"` // ignore this field as condition.
	}

	x, err := sqx.CreateSQL(`select a,b,c from t`, cond{})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where c2 > ? and e = ?`, x.Q)
	assert.Equal(t, []interface{}{0, ""}, x.Vars)

	x, err = sqx.CreateSQL(`select a,b,c from t`, cond{E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where c2 > ?`, x.Q)
	assert.Equal(t, []interface{}{0}, x.Vars)

	x, err = sqx.CreateSQL(`select a,b,c from t order by a`, &cond{E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where c2 > ? order by a`, x.Q)
	assert.Equal(t, []interface{}{0}, x.Vars)

	x, err = sqx.CreateSQL(`select a,b,c from t order by a`, cond{D: -1, E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a,b,c from t order by a`, x.Q)

	x, err = sqx.CreateSQL(`select a,b,c from t order by a`, cond{B: "bb", D: -1, E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where b = ? order by a`, x.Q)
	assert.Equal(t, []interface{}{"bb"}, x.Vars)

	x, err = sqx.CreateSQL(`select a,b,c from t where a = 1 order by a`, cond{B: "bb", D: -1, E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where a = 1 and b = ? order by a`, x.Q)
	assert.Equal(t, []interface{}{"bb"}, x.Vars)

	x, err = sqx.CreateSQL(`select a,b,c from t where a = 1 or a = 2 order by a`, cond{B: "bb", D: -1, E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where (a = 1 or a = 2) and b = ? order by a`, x.Q)
	assert.Equal(t, []interface{}{"bb"}, x.Vars)

	x, err = sqx.CreateSQL(`select a,b,c from t where a = 1 or a = 2 order by a`, cond{C: 10, D: -1, E: "null"})
	assert.Nil(t, err)
	assert.Equal(t, `select a, b, c from t where (a = 1 or a = 2) and c0 > ? order by a`, x.Q)
	assert.Equal(t, []interface{}{10}, x.Vars)

	x.Append(`limit ?,?`, 0, 100)
	assert.Equal(t, `select a, b, c from t where (a = 1 or a = 2) and c0 > ? order by a limit ?,?`, x.Q)
	assert.Equal(t, []interface{}{10, 0, 100}, x.Vars)
}
