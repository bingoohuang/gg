package sqx

import "database/sql"

func open(driverName, dataSourceName string) (*sql.DB, error) {
	driverName = DetectDriverName(driverName, dataSourceName)
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return db, err
	}

	RegisterDriverName(db.Driver(), driverName)
	return db, nil
}

func Open(driverName, dataSourceName string) (*Sqx, error) {
	db, err := open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return NewSqx(db), nil
}
