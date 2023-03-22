package sqlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountSqlx(t *testing.T) {
	countParsed, _ := Parse(`select count(*)`)
	parsed, err := Parse(`select a, b, c from t_mytable where a = ? order by b limit ?,?`)
	assert.Nil(t, err)

	selectQuery, ok := parsed.(*Select)
	assert.True(t, ok)

	selectQuery.SelectExprs = countParsed.(*Select).SelectExprs
	selectQuery.OrderBy = nil
	selectQuery.Having = nil
	selectQuery.Limit = nil
	assert.Equal(t, `select count(*) from t_mytable where a = ?`, String(selectQuery))
}
