package sqx

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"reflect"
	"strings"
	"sync"
)

var (
	sqlDriverNamesByType     = map[reflect.Type]string{}
	sqlDriverNamesByTypeLock sync.Mutex
	sqlDriverNamesByTypeOnce sync.Once
)

// The database/sql API doesn't provide a way to get the registry name for
// a driver from the driver type.
func sqlDriverToDriverName(driver driver.Driver) string {
	driverType := reflect.TypeOf(driver)
	if name, ok := sqlDriverNamesByType[driverType]; ok {
		return name
	}

	sqlDriverNamesByTypeOnce.Do(func() {
		for _, driverName := range sql.Drivers() {
			// Tested empty string DSN with MySQL, PostgreSQL, and SQLite3 drivers.
			if db, err := sql.Open(driverName, ""); err != nil {
				log.Printf("E! test empty dsn: %v", err)
			} else {
				sqlDriverNamesByType[reflect.TypeOf(db.Driver())] = driverName
			}
		}
	})

	return sqlDriverNamesByType[driverType]
}

// RegisterDriverName register the driver name for the current db.
func RegisterDriverName(d driver.Driver, driverName string) {
	sqlDriverNamesByTypeLock.Lock()
	defer sqlDriverNamesByTypeLock.Unlock()

	sqlDriverNamesByType[reflect.TypeOf(d)] = driverName
}

// DriverName returns the driver name for the current db.
func DriverName(d driver.Driver) string {
	sqlDriverNamesByTypeLock.Lock()
	defer sqlDriverNamesByTypeLock.Unlock()

	return sqlDriverToDriverName(d)
}

// DetectDriverName detects the driver name for database source name.
func DetectDriverName(driverName, dataSourceName string) string {
	// DB | driverName | DSN
	// ---|---|---
	// MySQL |mysql |user:pass@tcp(127.0.0.1:3306)/mydb?charset=utf8
	// 达梦 |dm| dm://user:pass@127.0.0.1:5236
	// 人大金仓|pgx|postgres://user:pass@127.0.0.1:54321/mydb?sslmode=disable
	// 华为GaussDB|opengauss://user:pass@127.0.0.1:54321/mydb?sslmode=disable

	if driverName == "" {
		if strings.Contains(dataSourceName, "://") {
			driverName = dataSourceName[:strings.Index(dataSourceName, "://")]
		} else {
			driverName = "mysql"
		}
	}

	switch driverName { // use pgx when dsn starts with postgres
	case "postgres":
		driverName = "pgx"
	}

	return driverName
}
