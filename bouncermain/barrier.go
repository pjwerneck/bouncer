package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type Barrier struct {
	Name    string
	Size    uint64
	mu      *sync.RWMutex
	waiting int64
	done    bool
	waitC   chan bool
	Stats   *Stats
}

var barriers = map[string]*Barrier{}
var barriersMutex = &sync.Mutex{}

func newBarrier(name string, size uint64) *Barrier {
	barrier := &Barrier{
		Name:  name,
		Size:  size,
		mu:    &sync.RWMutex{},
		Stats: &Stats{},
		waitC: make(chan bool),
	}
	barriers[name] = barrier
	return barrier
}

func getBarrier(name string, size uint64) (*Barrier, error) {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		barrier = newBarrier(name, size)
		logger.Infof("New barrier created: name=%v, size=%v", name, size)
	}
	return barrier, nil
}

func (b *Barrier) Wait(maxwait time.Duration) error {
	b.mu.Lock()
	if b.done {
		b.mu.Unlock()
		return ErrBarrierClosed
	}
	count := atomic.AddInt64(&b.waiting, 1)
	b.mu.Unlock()

	if count == int64(b.Size) {
		b.mu.Lock()
		close(b.waitC)
		b.done = true
		b.mu.Unlock()
		return nil
	}

	switch {
	case maxwait < 0:
		<-b.waitC
		return nil
	case maxwait == 0:
		select {
		case <-b.waitC:
			return nil
		default:
			atomic.AddInt64(&b.waiting, -1)
			return ErrTimedOut
		}
	default:
		timer := time.NewTimer(maxwait)
		defer timer.Stop()
		select {
		case <-b.waitC:
			return nil
		case <-timer.C:
			atomic.AddInt64(&b.waiting, -1)
			return ErrTimedOut
		}
	}
}

func deleteBarrier(name string) error {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		return ErrNotFound
	}

	barrier.mu.Lock()
	if !barrier.done {
		close(barrier.waitC)
		barrier.done = true
	}
	barrier.mu.Unlock()

	delete(barriers, name)
	return nil
}
