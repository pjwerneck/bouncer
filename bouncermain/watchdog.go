package bouncermain

import (
	"sync"
	"time"
)

type Watchdog struct {
	Name   string
	timer  *time.Timer
	mu     *sync.Mutex
	resetC chan bool
}

var watchdogs = map[string]*Watchdog{}
var watchdogsMutex = &sync.Mutex{}

func newWatchdog(name string, expires time.Duration) (watchdog *Watchdog) {
	watchdog = &Watchdog{
		Name:   name,
		timer:  time.NewTimer(expires),
		mu:     &sync.Mutex{},
		resetC: make(chan bool),
	}

	watchdogs[name] = watchdog

	go watchdog.watch()

	return watchdog
}

func (watchdog *Watchdog) watch() {
	select {
	case <-watchdog.timer.C:
		logger.Info("watchdog triggered")
	case <-watchdog.resetC:
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

	return nil
}

func (watchdog *Watchdog) GetStats() *Metrics {
	return nil
}
