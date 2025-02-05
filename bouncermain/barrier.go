package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type Barrier struct {
	Name       string
	Size       uint64
	mu         *sync.RWMutex
	waiting    int64
	generation int64
	current    waitGroup
	Stats      *Stats
}

type waitGroup struct {
	waitC chan bool
	doneC chan bool
	gen   int64
}

var barriers = map[string]*Barrier{}
var barriersMutex = &sync.Mutex{}

func newBarrier(name string, size uint64) *Barrier {
	barrier := &Barrier{
		Name:  name,
		Size:  size,
		mu:    &sync.RWMutex{},
		Stats: &Stats{},
		current: waitGroup{
			waitC: make(chan bool),
			doneC: make(chan bool),
			gen:   0,
		},
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
	count := atomic.AddInt64(&b.waiting, 1)
	currentGroup := b.current
	b.mu.Unlock()

	if count == int64(b.Size) {
		b.mu.Lock()
		if atomic.LoadInt64(&b.waiting) == int64(b.Size) {
			atomic.StoreInt64(&b.waiting, 0)
			close(currentGroup.waitC)
			close(currentGroup.doneC)
			b.current = waitGroup{
				waitC: make(chan bool),
				doneC: make(chan bool),
				gen:   currentGroup.gen + 1,
			}
		}
		b.mu.Unlock()
		return nil
	}

	switch {
	case maxwait < 0:
		select {
		case <-currentGroup.waitC:
			return nil
		case <-currentGroup.doneC:
			return nil
		}

	case maxwait == 0:
		select {
		case <-currentGroup.waitC:
			return nil
		case <-currentGroup.doneC:
			return nil
		default:
			atomic.AddInt64(&b.waiting, -1)
			return ErrTimedOut
		}

	default:
		timer := time.NewTimer(maxwait)
		defer timer.Stop()
		select {
		case <-currentGroup.waitC:
			return nil
		case <-currentGroup.doneC:
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
	close(barrier.current.waitC)
	close(barrier.current.doneC)
	barrier.mu.Unlock()

	delete(barriers, name)
	return nil
}
