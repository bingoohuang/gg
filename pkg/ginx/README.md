# ginx

some extensions for gin.

## ginpprof

[ginpprof](pkg/ginpprof/README.md) - A wrapper for [gin](https://github.com/gin-gonic/gin) to use `net/http/pprof` easily. 

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/bingoohuang/gg/pkg/ginx/ginpprof"
)

func main() {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	ginpprof.Wrap(router)

	// ginpprof also plays well with *gin.RouterGroup
	// group := router.Group("/debug/pprof")
	// ginpprof.WrapGroup(group)

	router.Run(":57047")
}
```

Now visit [http://127.0.0.1:57047/debug/pprof/](http://127.0.0.1:57047/debug/pprof/) and you'll see what you want.

More scripts:

1. `go tool pprof -http=:8080 http://127.0.0.1:57047/debug/pprof/profile`
1. `go tool pprof -http=:8080 http://127.0.0.1:57047/debug/pprof/heap`

## func adapter

The func adapter can adapt any prototype of router function to gin.HandlerFunc by registering in advance.

```go
import (
	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/gin-gonic/gin"
)

func main() {
	r := adapt.Adapt(gin.New())
	r.RegisterAdapter(func(f func(string) string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(http.StatusOK, f(StringArg(c)))
		}
	})

	// the binding func will be adapted to gin.HandlerFunc.
	r.GET("/user/:name", func(name string) string {
		return fmt.Sprintf("Hello %s", name)
	})

	// or use direct gin.HandlerFunc.
	r.GET("/direct/:name", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello Direct %s", c.Param("name"))
	})

	r.Run(":8080")
}

```

## anyfn adapter

```go
import (
	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/gintest"
	"github.com/gin-gonic/gin"
)

func main() {
	r := adapt.Adapt(gin.New())
	af := anyfn.NewAdapter()
	r.RegisterAdapter(af)

	// This handler will match /user/john but will not match /user/ or /user
	r.GET("/user/:name", func(name string) string {
		return fmt.Sprintf("Hello %s", name)
	})

	type MyObject struct {
		Name string
	}

    // af.F can adapt any function as you desired  to gin.HandlerFunc.
	r.POST("/MyObject1", af.F(func(m MyObject) string {
		return "Object: " + m.Name
	}))

	r.Run(":8080")
}
```

## hlog for logrus

```go
import (
	"database/sql"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"

	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/hlog"
	"github.com/gin-gonic/gin"
)

func init() {
	_, _ = golog.SetupLogrus(nil, "", "")
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	af := anyfn.NewAdapter()
	hf := hlog.NewAdapter(hlog.NewLogrusStore())
	r := adapt.Adapt(gin.New(), af, hf)
	r.Use(ginlogrus.Logger(nil, true))

	r.POST("/hello", af.F(func() string { return "Hello hello!" }), hf.F(hf.Biz("你好啊")))
	r.POST("/world", af.F(func() string { return "Hello world!" }))
	r.POST("/bye", af.F(func() string { return "Hello bye!" }), hf.F(hf.Ignore()))

	r.Run(":8080")
}
```

## hlog for mysql

```go
import (
	"database/sql"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"

	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/hlog"
	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	_, _ = golog.SetupLogrus(nil, "", "")
	gin.SetMode(gin.ReleaseMode)
}

const DSN = `root:root@tcp(127.0.0.1:3306)/httplog?charset=utf8mb4&parseTime=true&loc=Local`


func main() {
	db, err := sql.Open("mysql", DSN)
	assert.Nil(t, err)

	af := anyfn.NewAdapter()
	hf := hlog.NewAdapter(hlog.NewSQLStore(db, "biz_log"))
	r := adapt.Adapt(gin.New(), af, hf)
	r.Use(ginlogrus.Logger(nil, true))

	r.POST("/hello", af.F(handleIndex), hf.F(hf.Biz("回显处理hlog")))
	r.POST("/world", af.F(func() string { return "Hello world!" }), hf.F(hf.Biz("世界你好")))
	r.POST("/bye", af.F(func() string { return "Hello bye!" }), hf.F(hf.Ignore()))

	r.Run(":8080")
}
```

### Prepare log tables

业务日志表定义，根据具体业务需要，必须字段为主键`id`（名字固定）, 示例: [mysql](testdata/mysql.sql)

<details>
  <summary>
    <p>日志表建表规范</p>
  </summary>

字段注释包含| 或者字段名 | 说明
---|---|---
内置类:||
`httplog:"id"`|id| 日志记录ID
`httplog:"created"`|created| 创建时间
`httplog:"ip"` |ip|当前机器IP
`httplog:"addr"` |addr|http客户端地址
`httplog:"hostname"` |hostname|当前机器名称
`httplog:"pid"` |pid|应用程序PID
`httplog:"started"` |start|开始时间
`httplog:"end"` |end|结束时间
`httplog:"cost"` |cost|花费时间（ms)
`httplog:"biz"` |biz|业务名称，eg `httplog.Biz("项目列表")`
请求类:||
`httplog:"req_head_xxx"` |req_head_xxx|请求中的xxx头
`httplog:"req_heads"` |req_heads|请求中的所有头
`httplog:"req_method"` |req_method|请求method
`httplog:"req_url"` |req_url|请求URL
`httplog:"req_path_xxx"` |req_path_xxx|请求URL中的xxx路径参数
`httplog:"req_paths"` |req_paths|请求URL中的所有路径参数
`httplog:"req_query_xxx"` |req_query_xxx|请求URl中的xxx查询参数
`httplog:"req_queries"` |req_queries|请求URl中的所有查询参数
`httplog:"req_param_xxx"` |req_param_xxx|请求中query/form的xxx参数
`httplog:"req_params"` |req_params|请求中query/form的所有参数
`httplog:"req_body"` |req_body|请求体
`httplog:"req_json"` |req_json|请求体（当Content-Type为JSON时)
`httplog:"req_json_xxx"` |req_json_xxx|请求体JSON中的xxx属性
响应类:||
`httplog:"rsp_head_xxx"` |rsp_head_xxx|响应中的xxx头
`httplog:"rsp_heads"` |rsp_heads|响应中的所有头
`httplog:"rsp_body"` |rsp_body|响应体
`httplog:"rsp_json"` |rsp_json|响应体JSON（当Content-Type为JSON时)
`httplog:"rsp_json_xxx"`|rsp_json_xxx| 请求体JSON中的xxx属性
`httplog:"rsp_status"`|rsp_status| 响应编码
上下文:||
`httplog:"ctx_xxx"` |ctx_xxx|上下文对象xxx的值, 通过api设置: `hlog.PutAttr(c, "xxx", "yyy")` 或者 `hlog.PutAttrMap(r, hlog.Attrs{"name": "alice", "female": true})`, See [example](pkg/hlog/hlog_test.go#L78)
</details>
