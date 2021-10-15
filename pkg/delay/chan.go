package delay

import (
	"context"
	"sync"
	"time"
)

const defaultKey = "_default"

func NewChan(ctx context.Context, fn func(interface{}), delay time.Duration) *Chan {
	d := &Chan{fn: fn, Map: &sync.Map{}, wg: &sync.WaitGroup{}, stop: make(chan struct{}, 1)}
	d.Map.Store(defaultKey, make(chan interface{}, 1))
	d.wg.Add(1)
	go d.run(ctx, delay)
	return d
}

type Chan struct {
	fn   func(interface{})
	Map  *sync.Map
	wg   *sync.WaitGroup
	stop chan struct{}
}

func (c *Chan) run(ctx context.Context, delay time.Duration) {
	defer c.wg.Done()
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	defer c.consume()

	for {
		select {
		case <-ticker.C:
			c.consume()
		case <-c.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *Chan) Close() error {
	c.stop <- struct{}{}
	c.wg.Wait()
	return nil
}

func (c *Chan) consume() {
	c.Map.Range(func(k, value interface{}) bool {
		select {
		case v := <-value.(chan interface{}):
			c.fn(v)
		default:
		}
		return true
	})
}

func (c *Chan) PutKey(k string, v interface{}) {
	if ch, ok := c.Map.Load(k); ok {
		replace(ch.(chan interface{}), v)
		return
	}

	ch, _ := c.Map.LoadOrStore(k, make(chan interface{}, 1))
	replace(ch.(chan interface{}), v)
}

func (c *Chan) Put(v interface{}) { c.PutKey(defaultKey, v) }

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
