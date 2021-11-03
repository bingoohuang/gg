package cmap

import (
	"encoding/json"
	"sync"
)

// Map is a "thread" safe map of type string:Anything.
// To avoid lock bottlenecks this map is dived to several (ShardCount) map shards.
type Map struct {
	Shared     []*Shared
	ShardCount int
}

// Shared is a "thread" safe string to anything map.
type Shared struct {
	items        map[string]interface{}
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

type Option struct {
	ShardCount int
}

type OptionFn func(o *Option)

func WithShardCount(v int) OptionFn {
	return func(o *Option) {
		o.ShardCount = v
	}
}

// New creates a new concurrent map.
func New(options ...OptionFn) *Map {
	option := &Option{}
	for _, fn := range options {
		fn(option)
	}
	if option.ShardCount <= 0 {
		option.ShardCount = 32
	}

	m := make([]*Shared, option.ShardCount)
	for i := 0; i < option.ShardCount; i++ {
		m[i] = &Shared{items: make(map[string]interface{})}
	}
	return &Map{
		Shared:     m,
		ShardCount: option.ShardCount,
	}

}

// GetShard returns shard under given key
func (m Map) GetShard(key string) *Shared {
	return m.Shared[uint(fnv32(key))%uint(m.ShardCount)]
}

func (m Map) MSet(data map[string]interface{}) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

// Set sets the given value under the specified key.
func (m Map) Set(key string, value interface{}) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// UpsertCb is callback to return new element to be inserted into the map
// It is called while lock is held, therefore it MUST NOT
// try to access other keys in same map, as it can lead to deadlock since
// Go sync.RWLock is not reentrant
type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

// Upsert is Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m Map) Upsert(key string, value interface{}, cb UpsertCb) (res interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent sets the given value under the specified key if no value was associated with it.
func (m Map) SetIfAbsent(key string, value interface{}) bool {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

// GetString retrieves a string element from map under given key.
func (m Map) GetString(key string) (string, bool) {
	val, ok := m.Get(key)
	if !ok {
		return "", false
	}

	return val.(string), true
}

// Get retrieves an element from map under given key.
func (m Map) Get(key string) (interface{}, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map.
func (m Map) Count() int {
	count := 0
	for i := 0; i < m.ShardCount; i++ {
		shard := m.Shared[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has looks up an item under specified key.
func (m Map) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Del removes an element from the map.
func (m Map) Del(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCb is a callback executed in a map.RemoveCb() call, while Lock is held
// If returns true, the element will be removed from the map
type RemoveCb func(key string, v interface{}, exists bool) bool

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m Map) RemoveCb(key string, cb RemoveCb) bool {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// Pop removes an element from the map and returns it
func (m Map) Pop(key string) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

// IsEmpty checks if map is empty.
func (m Map) IsEmpty() bool {
	return m.Count() == 0
}

// Tuple is used by the Iter & IterBuffered functions to wrap two variables together over a channel.
type Tuple struct {
	Key string
	Val interface{}
}

// Iter returns a buffered iterator which could be used in a for range loop.
func (m Map) Iter() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

// Clear removes all items from map.
func (m Map) Clear() {
	for item := range m.Iter() {
		m.Del(item.Key)
	}
}

// Returns a array of channels that contains elements in each shard,
// which likely takes a snapshot of `m`.
// It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines.
func snapshot(m Map) (chans []chan Tuple) {
	chans = make([]chan Tuple, m.ShardCount)
	wg := sync.WaitGroup{}
	wg.Add(m.ShardCount)
	// Foreach shard.
	for index, shard := range m.Shared {
		go func(index int, shard *Shared) {
			// Foreach key, value pair.
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{Key: key, Val: val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// fanIn reads elements from channels `chans` into channel `out`.
func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

// Items returns all items as map[string]interface{}
func (m Map) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

// IterCb is iterator callback,called for every key,value found in
// maps. RLock is held for all calls for a given shard
// therefore callback sess consistent view of a shard,
// but not across the shards
type IterCb func(key string, v interface{})

// IterCb is callback based iterator, cheapest way to read
// all elements in a map.
func (m Map) IterCb(fn IterCb) {
	for idx := range m.Shared {
		shard := (m.Shared)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

// Keys returns all keys as []string
func (m Map) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(m.ShardCount)
		for _, shard := range m.Shared {
			go func(shard *Shared) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

// MarshalJSON reviles Map "private" variables to json marshal.
func (m Map) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
