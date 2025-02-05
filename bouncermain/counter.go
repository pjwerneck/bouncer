package bouncermain

import (
	"sync"
	"sync/atomic"
)

type Counter struct {
	Name  string
	value int64
	mutex *sync.RWMutex
	Stats *Stats
}

var counters = map[string]*Counter{}
var countersMutex = &sync.Mutex{}

func newCounter(name string) *Counter {
	counter := &Counter{
		Name:  name,
		mutex: &sync.RWMutex{},
		Stats: &Stats{},
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
		logger.Infof("New counter created: name=%v", name)
	}

	return counter, nil
}

func (c *Counter) Count(amount int64) int64 {
	return atomic.AddInt64(&c.value, amount)
}

func (c *Counter) Reset(value int64) {
	atomic.StoreInt64(&c.value, value)
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
