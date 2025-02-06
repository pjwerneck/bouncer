package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type CounterStats struct {
	Value      uint64 `json:"value"`
	Increments uint64 `json:"increments"`
	Resets     uint64 `json:"resets"`
	CreatedAt  string `json:"created_at"`
}

type Counter struct {
	Name  string
	value int64
	mutex *sync.RWMutex
	Stats *CounterStats
}

var counters = map[string]*Counter{}
var countersMutex = &sync.Mutex{}

func newCounter(name string) *Counter {
	counter := &Counter{
		Name:  name,
		mutex: &sync.RWMutex{},
		Stats: &CounterStats{CreatedAt: time.Now().Format(time.RFC3339)},
	}
	counters[name] = counter
	return counter
}

func getCounter(name string) (*Counter, error) {
	countersMutex.Lock()
	defer countersMutex.Unlock()

	counter, ok := counters[name]
	if !ok {
		counter = newCounter(name)
	}

	return counter, nil
}

func (c *Counter) Count(amount int64) int64 {
	val := atomic.AddInt64(&c.value, amount)
	atomic.AddUint64(&c.Stats.Value, 1)
	atomic.AddUint64(&c.Stats.Increments, 1)
	return val
}

func (c *Counter) Reset(value int64) {
	atomic.StoreInt64(&c.value, value)
	atomic.StoreUint64(&c.Stats.Value, 0)
	atomic.AddUint64(&c.Stats.Resets, 1)
}

func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

func deleteCounter(name string) error {
	countersMutex.Lock()
	defer countersMutex.Unlock()

	if _, ok := counters[name]; !ok {
		return ErrNotFound
	}

	delete(counters, name)
	return nil
}

func getCounterStats(name string) (*CounterStats, error) {
	countersMutex.Lock()
	defer countersMutex.Unlock()

	counter, ok := counters[name]
	if !ok {
		return nil, ErrNotFound
	}

	return counter.Stats, nil
}
