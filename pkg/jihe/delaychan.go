package jihe

import (
	"context"
	"sync"
	"time"
)

const defaultKey = "_default"

func NewDelayChan(ctx context.Context, fn func(interface{}), delay time.Duration) *DelayChan {
	d := &DelayChan{fn: fn, Map: &sync.Map{}, wg: &sync.WaitGroup{}}
	d.Map.Store(defaultKey, make(chan interface{}, 1))
	d.wg.Add(1)
	go d.run(ctx, delay)
	return d
}

type DelayChan struct {
	fn  func(interface{})
	Map *sync.Map
	wg  *sync.WaitGroup
}

func (c *DelayChan) run(ctx context.Context, delay time.Duration) {
	defer c.wg.Done()
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !c.consume() {
				return
			}
		case <-ctx.Done():
			c.consume()
			return
		}
	}
}

func (c *DelayChan) Close() error {
	c.Map.Range(func(_, value interface{}) bool {
		close(value.(chan interface{}))
		return true
	})
	c.wg.Wait()
	return nil
}

func (c *DelayChan) consume() bool {
	closeCount := 0
	count := 0
	c.Map.Range(func(k, value interface{}) bool {
		count++
		select {
		case v, ok := <-value.(chan interface{}):
			if ok {
				c.fn(v)
			} else {
				closeCount++
				c.Map.Delete(k)
			}
		default:
		}

		return true
	})

	return closeCount < count
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
