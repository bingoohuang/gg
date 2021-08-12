package hlog

import "github.com/sirupsen/logrus"

// LogrusStore stores the log as logurs info.
type LogrusStore struct{}

// NewLogrusStore returns a new LogrusStore.
func NewLogrusStore() *LogrusStore {
	return &LogrusStore{}
}

// Store stores the log in database like MySQL, InfluxDB, and etc.
func (s *LogrusStore) Store(log *Log) {
	logrus.Infof("http:%+v\n", log)
}

// Stores is the composite stores.
type Stores struct {
	Composite []Store
}

// Store stores the log in database like MySQL, InfluxDB, and etc.
func (s *Stores) Store(log *Log) {
	for _, v := range s.Composite {
		v.Store(log)
	}
}

// NewStores composes the stores as a Store.
func NewStores(stores ...Store) *Stores {
	return &Stores{Composite: stores}
}
