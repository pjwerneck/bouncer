package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type WatchdogStats struct {
	Waited    uint64 `json:"waited"`
	TimedOut  uint64 `json:"timed_out"`
	Kicks     uint64 `json:"kicks"`
	LastKick  string `json:"last_kick"`
	CreatedAt string `json:"created_at"`
}

type Watchdog struct {
	Name    string
	Stats   *WatchdogStats
	expires int64 // atomic unix nano when watchdog expires
}

var watchdogs = map[string]*Watchdog{}
var watchdogsMutex = &sync.RWMutex{}

func newWatchdog(name string, expires time.Duration) *Watchdog {
	now := time.Now()
	watchdog := &Watchdog{
		Name:    name,
		Stats:   &WatchdogStats{CreatedAt: now.Format(time.RFC3339)},
		expires: now.Add(expires).UnixNano(),
	}
	watchdogs[name] = watchdog
	return watchdog
}

func (w *Watchdog) Kick(expires time.Duration) error {
	atomic.StoreInt64(&w.expires, time.Now().Add(expires).UnixNano())
	atomic.AddUint64(&w.Stats.Kicks, 1)
	w.Stats.LastKick = time.Now().Format(time.RFC3339)
	return nil
}

func (w *Watchdog) Wait(maxwait time.Duration) error {
	deadline := time.Now().Add(maxwait)

	// Check if already expired
	if time.Now().UnixNano() >= atomic.LoadInt64(&w.expires) {
		atomic.AddUint64(&w.Stats.Waited, 1)
		return nil
	}

	for {
		// Check timeout
		if maxwait >= 0 && time.Now().After(deadline) {
			atomic.AddUint64(&w.Stats.TimedOut, 1)
			return ErrTimedOut
		}

		// Check expiration
		if time.Now().UnixNano() >= atomic.LoadInt64(&w.expires) {
			atomic.AddUint64(&w.Stats.Waited, 1)
			return nil
		}

		// Sleep until next check or deadline
		now := time.Now()
		sleepUntil := time.Unix(0, atomic.LoadInt64(&w.expires))
		if maxwait >= 0 && deadline.Before(sleepUntil) {
			sleepUntil = deadline
		}

		time.Sleep(min(sleepUntil.Sub(now), maxSleepDuration))
	}
}

func getWatchdog(name string, expires time.Duration) (watchdog *Watchdog, err error) {
	watchdogsMutex.RLock()
	watchdog, ok := watchdogs[name]
	watchdogsMutex.RUnlock()

	if ok {
		return watchdog, nil
	}

	// Watchdog doesn't exist, need to create it
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	// Check again in case another goroutine created it
	watchdog, ok = watchdogs[name]
	if ok {
		return watchdog, nil
	}

	watchdog = newWatchdog(name, expires)
	return watchdog, err
}

func getWatchdogStats(name string) (interface{}, error) {
	watchdogsMutex.RLock()
	defer watchdogsMutex.RUnlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		return nil, ErrNotFound
	}

	stats := &WatchdogStats{}
	*stats = *watchdog.Stats

	return stats, nil
}

func deleteWatchdog(name string) error {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	_, ok := watchdogs[name]
	if !ok {
		return ErrNotFound
	}

	delete(watchdogs, name)
	return nil
}
