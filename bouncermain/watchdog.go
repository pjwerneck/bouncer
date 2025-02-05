package bouncermain

import (
	"sync"
	"time"
)

type Watchdog struct {
	Name    string
	timer   *time.Timer
	mu      *sync.Mutex
	waitC   chan bool
	stopC   chan bool
	doneC   chan bool // New channel to signal goroutine completion
	expired bool
	Stats   *Metrics
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
		Stats: &Metrics{},
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
	return nil
}

func (watchdog *Watchdog) Wait(maxwait time.Duration) error {
	watchdog.mu.Lock()
	ch := watchdog.waitC // Get current channel
	expired := watchdog.expired
	watchdog.mu.Unlock()

	if expired {
		return nil // Return immediately if already expired
	}

	// Use RecvTimeout for consistent timeout handling across primitives
	_, err := RecvTimeout(ch, maxwait)
	if err != nil {
		return err
	}

	// Double check expiration after wait
	watchdog.mu.Lock()
	expired = watchdog.expired
	watchdog.mu.Unlock()

	if expired {
		return nil
	}
	return ErrTimedOut
}

func (watchdog *Watchdog) GetStats() *Metrics {
	return nil
}

func getWatchdogStats(name string) (stats *Metrics, err error) {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		return nil, ErrNotFound
	}

	return watchdog.Stats, nil
}
