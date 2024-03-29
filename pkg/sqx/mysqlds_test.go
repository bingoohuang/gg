package sqx_test

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/sqx"
	"github.com/stretchr/testify/assert"
)

func TestCompatibleDs(t *testing.T) {
	ds := sqx.CompatibleMySQLDs("MYSQL_PWD=8BE4 mysql -h 127.0.0.1 -P 9633 -u root")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs(`MYSQL_PWD="! 8BE4" mysql --host="127.0.0.1" -P 9633 -u 'root'`)
	assert.Equal(t, `root:! 8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("mysql -h 127.0.0.1 -P 9633 -u root -p8BE4")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("mysql -h 127.0.0.1 -P 9633 -u root -p8BE4 -Dtest")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/test?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("mysql -h127.0.0.1 -u root -p8BE4 -Dtest")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("127.0.0.1:9633 root/8BE4")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("127.0.0.1 root/8BE4")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("127.0.0.1:9633 root/8BE4 db=test")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/test?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("root:8BE4@127.0.0.1:9633/test")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:9633)/test?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("root:8BE4@127.0.0.1/test")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=Local`, ds)

	ds = sqx.CompatibleMySQLDs("root:8BE4@127.0.0.1")
	assert.Equal(t, `root:8BE4@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=true&loc=Local`, ds)
}
