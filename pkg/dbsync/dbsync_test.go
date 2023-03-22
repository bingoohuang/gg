package dbsync

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestSync(t *testing.T) {
	// create database db_test;
	// create table t_bucket (pk varchar(100) primary key, v int not null);
	// insert into t_bucket(pk, v) values('aa', 1);
	// insert into t_bucket(pk, v) values('bb', 1);
	// update t_bucket set v = v+1 where pk = 'aa';
	// update t_bucket set v = v+1 where pk = 'bb';
	// delete from t_bucket where pk = 'aa';
	// delete from t_bucket where pk = 'bb';
	dsn := "root:root@tcp(127.0.0.1:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	assert.Nil(t, err)
	dbsync := NewDbSync(db, "t_bucket", WithPk("pk"), WithV("v"), WithDuration("10s"),
		WithNotify(func(event Event, id, v string) {
			fmt.Printf("event:%s pk: %s, v:%s\n", event, id, v)
		}))

	dbsync.Start()

	// time.Sleep(1 * time.Hour)
	dbsync.Stop()
}
