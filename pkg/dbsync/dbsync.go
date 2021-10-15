package dbsync

import (
	"context"
	"database/sql"
	"github.com/bingoohuang/gg/pkg/mapp"
	"github.com/bingoohuang/gg/pkg/sqx"
	"log"
	"time"
)

type DbSync struct {
	db     *sql.DB
	table  string
	config *Config
	cache  map[string]string
	stop   chan struct{}
}

func NewDbSync(db *sql.DB, table string, options ...Option) *DbSync {
	return &DbSync{db: db, table: table, config: createConfig(options), stop: make(chan struct{})}
}

func WithPk(v string) Option                          { return func(c *Config) { c.pk = v } }
func WithV(v string) Option                           { return func(c *Config) { c.v = v } }
func WithDuration(v string) Option                    { return func(c *Config) { c.duration, _ = time.ParseDuration(v) } }
func WithNotify(f func(e Event, id, v string)) Option { return func(c *Config) { c.notify = f } }
func WithContext(v context.Context) Option            { return func(c *Config) { c.Context = v } }

func (s *DbSync) Start() {
	go s.loop()
}

func (s *DbSync) loop() {
	t := time.NewTicker(s.config.duration)
	defer t.Stop()

	s.cache = make(map[string]string)
	query := s.config.CreateQuery(s.table)

	s.sync(query)

	for {
		select {
		case <-t.C:
			s.sync(query)
		case <-s.config.Context.Done():
			return
		case <-s.stop:
			return
		}
	}
}

type row struct {
	Pk string
	V  string
}

func (s *DbSync) sync(query string) {
	var rows []row
	err := sqx.NewSQL(query).Query(s.db, &rows)
	if err != nil {
		log.Printf("E! failed to execute query: %s, err: %v", query, err)
		return
	}

	current := mapp.Clone(s.cache)

	for _, r := range rows {
		v, ok := s.cache[r.Pk]
		if !ok || v != r.V {
			s.cache[r.Pk] = r.V

			if !ok { // 不存在
				s.config.notify(EventCreate, r.Pk, r.V)
			} else { // 存在，但是v变更了
				s.config.notify(EventModify, r.Pk, r.V)
			}
		}

		if ok {
			delete(current, r.Pk)
		}
	}

	for k, v := range current {
		delete(s.cache, k)
		s.config.notify(EventDelete, k, v)
	}
}

func (s *DbSync) Stop() {
	s.stop <- struct{}{}
}

type Option func(*Config)

type Event int

const (
	EventCreate Event = iota
	EventDelete
	EventModify
)

type Config struct {
	context.Context
	v        string
	pk       string
	duration time.Duration
	notify   func(event Event, id, v string)
}

func (c Config) CreateQuery(t string) string {
	return "select " + c.pk + " as pk," + c.v + " as v from " + t
}

func createConfig(options []Option) *Config {
	c := &Config{}

	for _, f := range options {
		f(c)
	}

	if c.pk == "" {
		c.pk = "id"
	}
	if c.v == "" {
		c.v = "v"
	}
	if c.duration <= 0 {
		c.duration = 3 * time.Second
	}
	if c.notify == nil {
		c.notify = func(e Event, id, v string) {
			log.Printf("event:%s id: %s v:%s,", e, id, v)
		}
	}
	if c.Context == nil {
		c.Context = context.Background()
	}

	return c
}

func (e Event) String() string {
	switch e {
	case EventCreate:
		return "EventCreate"
	case EventDelete:
		return "EventDelete"
	case EventModify:
		return "EventModify"
	}

	return "Unknown"
}
