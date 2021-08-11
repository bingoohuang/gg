package sqx

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"sync"
)

var sqlDriverNamesByType = map[reflect.Type]string{}
var sqlDriverNamesByTypeOnce sync.Once

// The database/sql API doesn't provide a way to get the registry name for
// a driver from the driver type.
func sqlDriverToDriverName(driver driver.Driver) string {
	sqlDriverNamesByTypeOnce.Do(func() {
		for _, driverName := range sql.Drivers() {
			// Tested empty string DSN with MySQL, PostgreSQL, and SQLite3 drivers.
			if db, _ := sql.Open(driverName, ""); db != nil {
				sqlDriverNamesByType[reflect.TypeOf(db.Driver())] = driverName
			}
		}
	})

	return sqlDriverNamesByType[reflect.TypeOf(driver)]
}

// DriverName returns the driver name for the current db.
func DriverName(db *sql.DB) string { return sqlDriverToDriverName(db.Driver()) }
