# sqx

helper to generate sql on various conditions.

`go get github.com/bingoohuang/gg/pkg/sqx/...`

```go
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
// &{Query:select code, name, addr, email from person where name like ? and addr = ? order by code Vars:[%天问一号% 火星基地]}

ret, err = sqx.CreateSQL(`select code, name, addr, email from person order by code`, Cond{
	Name: "嫦娥",
	Addr: "广寒宫",
	Code: 100,
})
fmt.Println(fmt.Sprintf("%+v", ret), err)
// &{Query:select code, name, addr, email from person where name like ? and addr = ? and code > ? order by code Vars:[%嫦娥% 广寒宫 100]}

ret.Append("limit ?, ?", 1, 10) // 手动附加语句
fmt.Println(fmt.Sprintf("%+v", ret), err)
// &{Query:select code, name, addr, email from person where name like ? and addr = ? and code > ? order by code limit ?, ? Vars:[%嫦娥% 广寒宫 100 1 10]}
```

create counting sql:

```go
ret, err := sqx.SQL{
	Query: `select a, b, c from t where a = ? and b = ? order by a limit ?, ?`,
	Vars:  []interface{}{"地球", "亚洲", 0, 100},
}.CreateCount()
fmt.Println(fmt.Sprintf("%+v", ret), err) 
// &{Query:select count(*) from t where a = ? and b = ? Vars:[地球 亚洲]}
```

do update or query to database:

```go
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
```
