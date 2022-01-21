package sqlkv_test

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/gokv/sqlkv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"
	"github.com/src-d/go-mysql-server/memory"
	"github.com/src-d/go-mysql-server/server"
	"github.com/src-d/go-mysql-server/sql"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
)

func TestSQL(t *testing.T) {
	driver := sqle.NewDefault()
	db, err := createTestDatabase("testdb")
	assert.Nil(t, err)
	driver.AddDatabase(db)

	l, _ := net.Listen("tcp", ":0")
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("localhost:%d", port),
		Auth:     auth.NewNativeSingle("user", "pass", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(config, driver)
	assert.Nil(t, err)

	go func() {
		if err := s.Start(); err != nil {
			log.Print("start", err)
		}
	}()

	client, err := sqlkv.NewClient(sqlkv.Config{
		DataSourceName: fmt.Sprintf("user:pass@tcp(localhost:%d)/testdb", port),
		UpdateSQL:      "update kv set v = :v, updated = :time where k = :k and state = 1",
	})

	k := "Key1"
	assert.Nil(t, client.Set(k, "bingoohuang"))

	found, v, err := client.Get(k)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, "bingoohuang", v)

	err = client.Del(k)
	assert.Nil(t, err)

	found, v, err = client.Get(k)
	assert.Nil(t, err)
	assert.False(t, found)

	client.Get("Key2")
	client.Get("Key3")
}

func createTestDatabase(dbName string) (*memory.Database, error) {
	const tableName = "kv"

	db := memory.NewDatabase(dbName)
	table := memory.NewTable(tableName, sql.Schema{
		{Name: "k", Type: sql.VarChar(10), Nullable: false, Source: tableName, PrimaryKey: true},
		{Name: "v", Type: sql.Text, Nullable: false, Source: tableName},
		{Name: "state", Type: sql.Int8, Nullable: false, Source: tableName},
		{Name: "updated", Type: sql.VarChar(30), Nullable: true, Source: tableName},
		{Name: "created", Type: sql.VarChar(30), Nullable: true, Source: tableName},
	})

	fmt.Println(table.String())

	db.AddTable(tableName, table)
	ctx := sql.NewEmptyContext()

	rows := []sql.Row{
		sql.NewRow("Key1", `"value1"`, 1, nil, nil),
		sql.NewRow("Key2", `"value2"`, 1, nil, nil),
		sql.NewRow("Key3", `"value3"`, 1, nil, nil),
	}

	for _, row := range rows {
		if err := table.Insert(ctx, row); err != nil {
			return nil, err
		}
	}

	return db, nil
}
