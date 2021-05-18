package sqx_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/sqx"
	"github.com/stretchr/testify/assert"
)

func ExampleCreateCount() { // nolint:govet
	ret, err := sqx.SQL{
		Query: `select a, b, c from t where a = ? and b = ? order by a limit ?, ?`,
		Vars:  []interface{}{"地球", "亚洲", 0, 100},
	}.CreateCount()
	fmt.Println(fmt.Sprintf("%+v", ret), err)

	// Output:
	// &{Query:select count(*) from t where a = ? and b = ? Vars:[地球 亚洲] Ctx:<nil> Log:false} <nil>
}

func TestCreateCount(t *testing.T) {
	s := &sqx.SQL{
		Query: `select a,b,c from t order by a`,
	}

	c, err := s.CreateCount()
	assert.Nil(t, err)
	assert.Equal(t, `select count(*) from t`, c.Query)
	assert.Nil(t, c.Vars)
}
