package sqx

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"reflect"
	"sync"
)

var sqlDriverNamesByType = map[reflect.Type]string{}
var sqlDriverNamesByTypeLock sync.Mutex
var sqlDriverNamesByTypeOnce sync.Once

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
