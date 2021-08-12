package sqlrun_test

import (
	"database/sql"
	"testing"

	"github.com/bingoohuang/gg/pkg/ginx/sqlrun"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

const DSN = `root:root@tcp(127.0.0.1:3306)/httplog?charset=utf8mb4&parseTime=true&loc=Local`

func TestSQLQuery(t *testing.T) {
	db, err := sql.Open("mysql", DSN)
	assert.Nil(t, err)

	run := sqlrun.NewSQLRun(db, sqlrun.NewMapPreparer(""))
	result := run.DoQuery("select 1")
	assert.Nil(t, result.Error)
	assert.Equal(t, [][]string{{"1"}}, result.Rows)
	assert.Equal(t, []string{"1"}, result.Headers)
	assert.True(t, result.IsQuery)
}
