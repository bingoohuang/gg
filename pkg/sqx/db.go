package sqx

import (
	"database/sql"
	"net/url"
	"strings"
)

func open(driverName, dataSourceName string) (*sql.DB, error) {
	driverName = DetectDriverName(driverName, dataSourceName)
	dataSourceName = tryUrlEncodePass(dataSourceName)
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return db, err
	}

	RegisterDriverName(db.Driver(), driverName)
	return db, nil
}

func tryUrlEncodePass(dataSourceName string) string {
	s := dataSourceName
	p1 := strings.Index(s, "://")
	p2 := 0
	p3 := 0
	if p1 > 0 {
		s = s[p1+3:]
		p2 = strings.Index(s, ":")
	}
	if p2 > 0 {
		s = s[p2+1:]
		p3 = strings.LastIndex(s, "@")
	}
	if p3 > 0 {
		pp0 := s[:p3]
		if pp1 := url.QueryEscape(pp0); pp0 != pp1 {
			dataSourceName = strings.Replace(dataSourceName, ":"+pp0+"@", ":"+pp1+"@", 1)
		}
	}
	return dataSourceName
}

func Open(driverName, dataSourceName string) (*sql.DB, *Sqx, error) {
	db, err := open(driverName, dataSourceName)
	if err != nil {
		return nil, nil, err
	}

	return db, NewSqx(db), nil
}
