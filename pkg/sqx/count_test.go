package sqx_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/gg/pkg/sqx"
	"github.com/stretchr/testify/assert"
)

func ExampleCreateCount() {
	ret, err := sqx.SQL{
		Q:    `select a, b, c from t where a = ? and b = ? order by a limit ?, ?`,
		Vars: []interface{}{"地球", "亚洲", 0, 100},
	}.CreateCount()
	fmt.Println(fmt.Sprintf("%+v", ret), err)

	// Output:
	// &{Name: Q:select count(*) from t where a = ? and b = ? Vars:[地球 亚洲] Ctx:<nil> NoLog:false Timeout:0s Limit:0 ConvertOptions:[]} <nil>
}

func TestCreateCount(t *testing.T) {
	s := &sqx.SQL{
		Q: `select a,b,c from t order by a`,
	}

	c, err := s.CreateCount()
	assert.Nil(t, err)
	assert.Equal(t, `select count(*) from t`, c.Q)
	assert.Nil(t, c.Vars)
}
