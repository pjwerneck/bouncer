package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type WatchdogStats struct {
	Waited          uint64  `json:"waited"`
	TimedOut        uint64  `json:"timed_out"`
	Kicks           uint64  `json:"kicks"`
	Triggered       uint64  `json:"triggered"`
	TotalWaitTime   uint64  `json:"total_wait_time"`
	AverageWaitTime float64 `json:"average_wait_time"`
	LastKick        string  `json:"last_kick"`
	CreatedAt       string  `json:"created_at"`
}

type Watchdog struct {
	Name    string
	timer   *time.Timer
	mu      *sync.Mutex
	waitC   chan bool
	stopC   chan bool
	doneC   chan bool // New channel to signal goroutine completion
	expired bool
	Stats   *WatchdogStats
}

var watchdogs = map[string]*Watchdog{}
var watchdogsMutex = &sync.Mutex{}

func newWatchdog(name string, expires time.Duration) (watchdog *Watchdog) {
	watchdog = &Watchdog{
		Name:  name,
		timer: time.NewTimer(expires),
		mu:    &sync.Mutex{},
		waitC: make(chan bool),
		stopC: make(chan bool),
		doneC: make(chan bool),
		Stats: &WatchdogStats{CreatedAt: time.Now().Format(time.RFC3339)},
	}

	watchdogs[name] = watchdog
	go watchdog.watch()
	return watchdog
}

func (watchdog *Watchdog) watch() {
	defer func() {
		watchdog.doneC <- true // Signal completion
	}()

	for {
		select {
		case <-watchdog.timer.C:
			watchdog.mu.Lock()
			if !watchdog.expired {
				watchdog.expired = true
				close(watchdog.waitC)
				watchdog.waitC = make(chan bool)
			}
			watchdog.mu.Unlock()
		case <-watchdog.stopC:
			return
		}
	}
}

func (watchdog *Watchdog) reset(expires time.Duration) {
	watchdog.mu.Lock()
	defer watchdog.mu.Unlock()

	// Stop existing watch goroutine and wait for completion
	close(watchdog.stopC)
	<-watchdog.doneC

	// Stop and drain timer
	if !watchdog.timer.Stop() {
		select {
		case <-watchdog.timer.C:
		default:
		}
	}

	// Reset all channels and state
	watchdog.stopC = make(chan bool)
	watchdog.doneC = make(chan bool)
	watchdog.waitC = make(chan bool)
	watchdog.expired = false

	// Reset timer
	watchdog.timer.Reset(expires)

	// Start new watch goroutine
	go watchdog.watch()
}

func getWatchdog(name string, expires time.Duration) (watchdog *Watchdog, err error) {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		watchdog = newWatchdog(name, expires)
		logger.Infof("New watchdog created: %v", name)
	}

	return watchdog, err
}

func (watchdog *Watchdog) Kick(expires time.Duration) error {
	watchdog.reset(expires)
	atomic.AddUint64(&watchdog.Stats.Kicks, 1)
	watchdog.Stats.LastKick = time.Now().Format(time.RFC3339)
	return nil
}

func (watchdog *Watchdog) Wait(maxwait time.Duration) error {
	started := time.Now()
	watchdog.mu.Lock()
	ch := watchdog.waitC // Get current channel
	expired := watchdog.expired
	watchdog.mu.Unlock()

	if expired {
		atomic.AddUint64(&watchdog.Stats.Waited, 1)
		atomic.AddUint64(&watchdog.Stats.Triggered, 1)
		return nil // Return immediately if already expired
	}

	// Use RecvTimeout for consistent timeout handling across primitives
	_, err := RecvTimeout(ch, maxwait)
	if err != nil {
		atomic.AddUint64(&watchdog.Stats.TimedOut, 1)
		return err
	}

	// Update stats
	wait := uint64(time.Since(started) / time.Millisecond)
	atomic.AddUint64(&watchdog.Stats.Waited, 1)
	atomic.AddUint64(&watchdog.Stats.TotalWaitTime, wait)
	atomic.AddUint64(&watchdog.Stats.Triggered, 1)

	return nil
}

func getWatchdogStats(name string) (stats *WatchdogStats, err error) {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		return nil, ErrNotFound
	}

	// Create a copy and calculate average
	stats = &WatchdogStats{}
	*stats = *watchdog.Stats
	waited := atomic.LoadUint64(&watchdog.Stats.Waited)
	if waited > 0 {
		stats.AverageWaitTime = float64(atomic.LoadUint64(&watchdog.Stats.TotalWaitTime)) / float64(waited)
	}

	return stats, nil
}

func deleteWatchdog(name string) error {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		return ErrNotFound
	}

	// Stop the watch goroutine
	close(watchdog.stopC)
	<-watchdog.doneC

	delete(watchdogs, name)
	return nil
}
