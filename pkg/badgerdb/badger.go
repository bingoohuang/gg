package badgerdb

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
	"io/ioutil"
	"time"
)

type Badger struct {
	DB *badger.DB
}

func WithInMemory(v bool) OpenOptionsFn { return func(o *OpenOptions) { o.InMemory = v } }
func WithPath(v string) OpenOptionsFn   { return func(o *OpenOptions) { o.Path = v } }

func Open(fns ...OpenOptionsFn) (*Badger, error) {
	db, err := badger.Open(OpenOptionsFns(fns).Create().Apply())
	if err != nil {
		return nil, err
	}

	return &Badger{DB: db}, nil
}

func (b *Badger) Close() error { return b.DB.Close() }

func WithWalkStart(v []byte) WalkOptionsFn { return func(o *WalkOptions) { o.Start = v } }
func WithMax(v int) WalkOptionsFn          { return func(o *WalkOptions) { o.Max = v } }
func WithPrefix(v []byte) WalkOptionsFn    { return func(o *WalkOptions) { o.Prefix = v } }
func WithReverse(v bool) WalkOptionsFn     { return func(o *WalkOptions) { o.Reverse = v } }
func WithOnlyKeys(v bool) WalkOptionsFn    { return func(o *WalkOptions) { o.OnlyKeys = v } }

func (b *Badger) Walk(f func(k, v []byte) error, fns ...WalkOptionsFn) error {
	wo := WalkOptionsFns(fns).Create()
	return b.DB.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(wo.NewIteratorOptions())
		defer it.Close()

		num := 0

		for it.Seek(wo.Start); it.ValidForPrefix(wo.Prefix) && (wo.Max <= 0 || num < wo.Max); it.Next() {
			item := it.Item()
			var v []byte

			if !wo.OnlyKeys {
				v, err = item.ValueCopy(nil)
				if err != nil {
					return err
				}
			}

			if err := f(item.Key(), v); err != nil {
				return err
			}
			num++
		}
		return nil
	})
}

func (b *Badger) Get(k []byte) (val []byte, er error) {
	er = b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}

		val, err = item.ValueCopy(nil)
		return err
	})

	return
}

func WithTTL(v time.Duration) SetOptionsFn { return func(o *SetOptions) { o.TTL = v } }
func WithMeta(v byte) SetOptionsFn         { return func(o *SetOptions) { o.Meta = v } }

func (b *Badger) Set(k, v []byte, fns ...SetOptionsFn) error {
	e := badger.NewEntry(k, v)
	SetOptionsFns(fns).Create().Apply(e)
	return b.DB.Update(func(txn *badger.Txn) error { return txn.SetEntry(e) })
}

type OpenOptions struct {
	InMemory bool
	Path     string
}

func (o OpenOptions) Apply() badger.Options {
	if o.InMemory {
		options := badger.DefaultOptions("").WithInMemory(true)
		options.Logger = nil
		return options
	}

	path := o.Path
	if path == "" {
		dir, err := ioutil.TempDir("", "badgerdb")
		if err != nil {
			panic(err)
		}
		path = dir
	}

	options := badger.DefaultOptions(path)
	options.Logger = nil
	return options
}

type OpenOptionsFn func(*OpenOptions)
type OpenOptionsFns []OpenOptionsFn

func (fns OpenOptionsFns) Create() *OpenOptions {
	o := &OpenOptions{}
	for _, f := range fns {
		f(o)
	}

	return o
}

type SetOptions struct {
	TTL  time.Duration
	Meta byte
}

func (o SetOptions) Apply(e *badger.Entry) {
	if o.TTL > 0 {
		e.WithTTL(o.TTL)
	}

	e.WithMeta(o.Meta)
}

type SetOptionsFn func(*SetOptions)
type SetOptionsFns []SetOptionsFn

func (fns SetOptionsFns) Create() *SetOptions {
	o := &SetOptions{}
	for _, f := range fns {
		f(o)
	}
	return o
}

type WalkOptions struct {
	Start    []byte
	Max      int
	OnlyKeys bool
	Reverse  bool
	Prefix   []byte
}

func (o WalkOptions) NewIteratorOptions() badger.IteratorOptions {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	opts.PrefetchValues = !o.OnlyKeys
	opts.Reverse = o.Reverse
	opts.Prefix = o.Prefix

	return opts
}

type WalkOptionsFn func(*WalkOptions)

type WalkOptionsFns []WalkOptionsFn

func (fns WalkOptionsFns) Create() *WalkOptions {
	o := &WalkOptions{}
	for _, fn := range fns {
		fn(o)
	}

	return o
}
