package hlog_test

import (
	"database/sql"
	"github.com/bingoohuang/gg/pkg/ginx/adapt"
	"github.com/bingoohuang/gg/pkg/ginx/anyfn"
	"github.com/bingoohuang/gg/pkg/ginx/gintest"
	"github.com/bingoohuang/gg/pkg/ginx/hlog"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	golog.Setup()
	gin.SetMode(gin.ReleaseMode)
}

func TestLogger(t *testing.T) {
	af := anyfn.NewAdapter()
	hf := hlog.NewAdapter(hlog.NewLogrusStore())
	r := adapt.Adapt(gin.New(), af, hf)
	r.Use(ginlogrus.Logger(nil, true))

	r.POST("/hello", af.F(func() string { return "Hello hello!" }), hf.F(hf.Biz("你好啊")))
	r.POST("/world", af.F(func() string { return "Hello world!" }))
	r.POST("/bye", af.F(func() string { return "Hello bye!" }), hf.F(hf.Ignore()))

	// r.Run(":8080")

	rr := gintest.Post("/hello", r)
	assert.Equal(t, "Hello hello!", rr.Body())

	rr = gintest.Post("/world", r)
	assert.Equal(t, "Hello world!", rr.Body())

	rr = gintest.Post("/bye", r)
	assert.Equal(t, "Hello bye!", rr.Body())
}

const DSN = `root:root@tcp(127.0.0.1:3306)/httplog?charset=utf8mb4&parseTime=true&loc=Local`

func TestNewSQLStore(t *testing.T) {
	db, err := sql.Open("mysql", DSN)
	assert.Nil(t, err)

	af := anyfn.NewAdapter()
	hf := hlog.NewAdapter(hlog.NewSQLStore(db, "biz_log"))
	r := adapt.Adapt(gin.New(), af, hf)
	r.Use(ginlogrus.Logger(nil, true))
	r.Use(func(c *gin.Context) {
		hlog.PutAttr(c, "age", 5000)
	})

	r.POST("/hello", af.F(handleIndex), hf.F(hf.Biz("回显处理hlog")))
	r.POST("/world", af.F(func() string { return "Hello world!" }), hf.F(hf.Biz("世界你好")))

	rr := gintest.Post("/hello", r, gintest.JSONVar(`{"name":"dingding"}`))
	assert.Equal(t, `{"name":"dingding"}`, rr.Body())

	gintest.Post("/world", r)
}

// simplest possible server that returns url as plain text.
func handleIndex(w http.ResponseWriter, r *http.Request, c *gin.Context) {
	// msg := fmt.Sprintf("You've called url %s", r.URL.String())
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK) // 200

	hlog.PutAttr(c, "xxx", "yyy")
	hlog.PutAttrMap(c, hlog.Attrs{"name": "alice", "female": true})

	var bytes []byte

	if r.Body != nil {
		bytes, _ = ioutil.ReadAll(r.Body)
	} else {
		bytes = []byte(`empty request body`)
	}

	_, _ = w.Write(bytes)
}
