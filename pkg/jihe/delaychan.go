package jihe

import (
	"context"
	"sync"
	"time"
)

const defaultKey = "_default"

func NewDelayChan(ctx context.Context, fn func(interface{}), delay time.Duration) *DelayChan {
	d := &DelayChan{fn: fn, Map: &sync.Map{}}
	d.Map.Store(defaultKey, make(chan interface{}, 1))
	go d.run(ctx, delay)
	return d
}

type DelayChan struct {
	fn  func(interface{})
	Map *sync.Map
}

func (c *DelayChan) run(ctx context.Context, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.consume()
		case <-ctx.Done():
			c.consume()
			return
		}
	}
}

func (c *DelayChan) consume() {
	c.Map.Range(func(_, value interface{}) bool {
		select {
		case v := <-value.(chan interface{}):
			c.fn(v)
		default:
		}

		return true
	})

}

func (c *DelayChan) PutKey(k string, v interface{}) {
	if ch, ok := c.Map.Load(k); ok {
		replace(ch.(chan interface{}), v)
		return
	}

	ch, _ := c.Map.LoadOrStore(k, make(chan interface{}, 1))
	replace(ch.(chan interface{}), v)
}

func (c *DelayChan) Put(v interface{}) { c.PutKey(defaultKey, v) }

func replace(ch chan interface{}, v interface{}) {
	// try to remove old one.
	select {
	case <-ch:
	default:
	}

	// try to replace the new one.
	select {
	case ch <- v:
	default:
	}
}
