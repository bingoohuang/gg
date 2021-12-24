package sqx

import "database/sql"

func Open(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return db, err
	}

	RegisterDriverName(db.Driver(), driverName)
	return db, nil
}

func OpenSqx(driverName, dataSourceName string) (*Sqx, error) {
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return NewSqx(db), nil
}
