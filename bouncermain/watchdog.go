package bouncermain

import (
	"sync"
	//"sync/atomic"
	"time"
)

type Watchdog struct {
	Name  string
	timer *time.Timer
	mu    *sync.Mutex

	resetC  chan bool
	trigger chan bool
}

var watchdogs = map[string]*Watchdog{}
var watchdogsMutex = &sync.Mutex{}

func newWatchdog(name string, interval uint64) (watchdog *Watchdog) {
	watchdog = &Watchdog{
		Name:  name,
		timer: time.NewTimer(time.Millisecond * time.Duration(interval)),
		mu:    &sync.Mutex{},
	}

	watchdogs[name] = watchdog

	go watchdog.watch()
	//go watchdog.Stats.Run()

	return watchdog
}

func (watchdog *Watchdog) watch() {
	<-watchdog.timer.C
	logger.Info("watchdog triggered")
}

func (watchdog *Watchdog) reset(interval uint64) {
	watchdog.timer.Stop()
	watchdog.timer.Reset(time.Millisecond * time.Duration(interval))
	logger.Debug("timer reset")

	go watchdog.watch()

}

func getWatchdog(name string, interval uint64) (watchdog *Watchdog) {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		watchdog = newWatchdog(name, interval)
		logger.Infof("New watchdog created: %v", name)
	}

	return watchdog
}

func (watchdog *Watchdog) Kick(interval uint64) (err error) {
	watchdog.mu.Lock()
	defer watchdog.mu.Unlock()

	// reset the timer
	watchdog.reset(interval)

	return nil
}

func (watchdog *Watchdog) GetStats() *Metrics {
	return nil
}
