module github.com/bingoohuang/gg

go 1.16

require (
	github.com/dgraph-io/badger/v3 v3.2103.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/goccy/go-yaml v1.8.10
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/juju/testing v0.0.0-20201216035041-2be42bba85f3 // indirect
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/pbnjay/pixfont v0.0.0-20200714042608-33b744692567
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	golang.org/x/text v0.3.5
	gorm.io/driver/mysql v1.1.0
	gorm.io/gorm v1.21.11
)

replace github.com/goccy/go-yaml => github.com/bingoohuang/go-yaml v1.8.11-0.20210719040622-7e6a9879a76a
