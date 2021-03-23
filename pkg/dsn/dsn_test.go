package dsn_test

import (
	"github.com/bingoohuang/gg/pkg/dsn"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseFlag(t *testing.T) {
	data := []struct {
		Input string
		Flag  dsn.Flag
		Fail  bool
	}{
		{Input: "user:pass@host", Flag: dsn.Flag{Username: "user", Password: "pass", Host: "host", Port: 0, Database: ""}},
		{Input: "user@host", Flag: dsn.Flag{Username: "user", Password: "", Host: "host", Port: 0, Database: ""}},
		{Input: "user/pass@host:3306", Flag: dsn.Flag{Username: "user", Password: "pass", Host: "host", Port: 3306, Database: ""}},
		{Input: "user/pass@host:3306/", Flag: dsn.Flag{Username: "user", Password: "pass", Host: "host", Port: 3306, Database: ""}},
		{Input: "user/pass@host:3306/mydb", Flag: dsn.Flag{Username: "user", Password: "pass", Host: "host", Port: 3306, Database: "mydb"}},
		{Input: "user:pass@host", Flag: dsn.Flag{Username: "user", Password: "pass", Host: "host", Port: 0, Database: ""}},
		{Input: "user:pass@host:3306/mydb", Flag: dsn.Flag{Username: "user", Password: "pass", Host: "host", Port: 3306, Database: "mydb"}},
		{Input: "user:p1:x2@y3@host:3306/mydb", Flag: dsn.Flag{Username: "user", Password: "p1:x2@y3", Host: "host", Port: 3306, Database: "mydb"}},

		{Input: "user/pass", Flag: dsn.Flag{}, Fail: true},
		{Input: "user/pass@", Flag: dsn.Flag{}, Fail: true},
		{Input: "user/pass@host:badport", Flag: dsn.Flag{}, Fail: true},
	}

	for _, dat := range data {
		f, err := dsn.ParseFlag(dat.Input)
		assert.Equal(t, err != nil, dat.Fail)
		if err == nil {
			assert.Equal(t, *f, dat.Flag)
		}
	}
}
