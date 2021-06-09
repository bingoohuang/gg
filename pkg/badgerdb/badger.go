package badgerdb

import (
	"encoding/binary"
	"errors"
	"github.com/dgraph-io/badger/v3"
	"time"
)

type Badger struct {
	DB *badger.DB
}

func New(path string, inMemory bool) (*Badger, error) {
	if inMemory {
		path = ""
	}
	options := badger.DefaultOptions(path).WithInMemory(inMemory)
	options.Logger = nil
	db, err := badger.Open(options)
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
	opts := wo.NewIteratorOptions()
	return b.DB.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(opts)
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
	so := SetOptionsFns(fns).Create()
	e := badger.NewEntry(k, v)
	so.Apply(e)
	return b.DB.Update(func(txn *badger.Txn) error { return txn.SetEntry(e) })
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

func Uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
