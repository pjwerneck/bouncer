package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type BarrierStats struct {
	Waiting         uint64  `json:"waiting"`
	Size            uint64  `json:"size"`
	TotalWaited     uint64  `json:"total_waited"`
	TimedOut        uint64  `json:"timed_out"`
	Triggered       uint64  `json:"triggered"`
	TotalWaitTime   uint64  `json:"total_wait_time"`
	CreatedAt       string  `json:"created_at"`
	AverageWaitTime float64 `json:"average_wait_time"`
}

type Barrier struct {
	Name    string
	Size    uint64
	mu      *sync.RWMutex
	waiting int64
	done    bool
	waitC   chan bool
	Stats   *BarrierStats
}

var barriers = map[string]*Barrier{}
var barriersMutex = &sync.RWMutex{}

func newBarrier(name string, size uint64) *Barrier {
	barrier := &Barrier{
		Name: name,
		Size: size,
		mu:   &sync.RWMutex{},
		Stats: &BarrierStats{
			CreatedAt: time.Now().Format(time.RFC3339),
			Size:      size,
		},
		waitC: make(chan bool),
	}
	barriers[name] = barrier
	return barrier
}

func getBarrier(name string, size uint64) (*Barrier, error) {
	barriersMutex.RLock()
	barrier, ok := barriers[name]
	barriersMutex.RUnlock()

	if ok {
		return barrier, nil
	}

	// Barrier doesn't exist, need to create it
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	// Check again in case another goroutine created it
	barrier, ok = barriers[name]
	if ok {
		return barrier, nil
	}

	barrier = newBarrier(name, size)
	return barrier, nil
}

func (b *Barrier) Wait(maxwait time.Duration) error {
	started := time.Now()
	atomic.AddUint64(&b.Stats.Waiting, 1)
	defer atomic.AddUint64(&b.Stats.Waiting, ^uint64(0)) // decrement

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

		// Update stats before returning
		atomic.AddUint64(&b.Stats.TotalWaited, 1)
		atomic.AddUint64(&b.Stats.Triggered, 1)
		wait := uint64(time.Since(started) / time.Millisecond)
		atomic.AddUint64(&b.Stats.TotalWaitTime, wait)
		return nil
	}

	var err error
	switch {
	case maxwait < 0:
		<-b.waitC
	case maxwait == 0:
		select {
		case <-b.waitC:
		default:
			atomic.AddInt64(&b.waiting, -1)
			atomic.AddUint64(&b.Stats.TimedOut, 1)
			return ErrTimedOut
		}
	default:
		timer := time.NewTimer(maxwait)
		defer timer.Stop()
		select {
		case <-b.waitC:
		case <-timer.C:
			atomic.AddInt64(&b.waiting, -1)
			atomic.AddUint64(&b.Stats.TimedOut, 1)
			return ErrTimedOut
		}
	}

	// Update stats before returning
	atomic.AddUint64(&b.Stats.TotalWaited, 1)
	wait := uint64(time.Since(started) / time.Millisecond)
	atomic.AddUint64(&b.Stats.TotalWaitTime, wait)

	return err
}

func deleteBarrier(name string) error {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		return ErrNotFound
	}

	// Close barrier channel and remove from map
	barrier.mu.Lock()
	if !barrier.done {
		close(barrier.waitC)
		barrier.done = true
	}
	barrier.mu.Unlock()

	delete(barriers, name)
	return nil
}

func getBarrierStats(name string) (interface{}, error) {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		return nil, ErrNotFound
	}

	// Create a copy of stats and calculate average
	stats := *barrier.Stats
	totalWaited := atomic.LoadUint64(&barrier.Stats.TotalWaited)
	if totalWaited > 0 {
		stats.AverageWaitTime = float64(atomic.LoadUint64(&barrier.Stats.TotalWaitTime)) / float64(totalWaited)
	}

	return &stats, nil
}
