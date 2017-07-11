package bouncermain

import (
	"sync"
	"time"
)

type Watchdog struct {
	Name   string
	timer  *time.Timer
	mu     *sync.Mutex
	waitC  chan bool
	resetC chan bool
	Stats  *Metrics
}

var watchdogs = map[string]*Watchdog{}
var watchdogsMutex = &sync.Mutex{}

func newWatchdog(name string, expires time.Duration) (watchdog *Watchdog) {
	watchdog = &Watchdog{
		Name:   name,
		timer:  time.NewTimer(expires),
		mu:     &sync.Mutex{},
		resetC: make(chan bool),
		waitC:  make(chan bool),
	}

	watchdogs[name] = watchdog

	go watchdog.watch()

	return watchdog
}

func (watchdog *Watchdog) watch() {
	select {
	case <-watchdog.timer.C:
		// broadcast by closing waitC, and replacing it immediately
		watchdog.mu.Lock()
		close(watchdog.waitC)
		watchdog.waitC = make(chan bool)
		watchdog.mu.Unlock()
	case <-watchdog.resetC:
		// do nothing
	}
}

func (watchdog *Watchdog) reset(expires time.Duration) {
	if watchdog.timer.Stop() {
		watchdog.resetC <- true
	}
	watchdog.timer.Reset(expires)

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

func (watchdog *Watchdog) Kick(expires time.Duration) (err error) {
	watchdog.mu.Lock()
	defer watchdog.mu.Unlock()

	// reset the timer
	watchdog.reset(expires)

	return nil
}

func (watchdog *Watchdog) Wait(maxwait time.Duration) (err error) {
	_, err = RecvTimeout(watchdog.waitC, maxwait)
	if err != nil {
		return err
	}

	return nil
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
