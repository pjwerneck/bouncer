package bouncermain

import (
	"sync"
	"time"
)

var mutexes = map[string]*Semaphore{}
var mutexesMutex = &sync.RWMutex{}

func newMutex(Name string) (semaphore *Semaphore) {
	mutexesMutex.Lock()
	defer mutexesMutex.Unlock()

	semaphore = &Semaphore{
		Name:     Name,
		Size:     uint64(1),
		Keys:     make(map[string]uint64),
		acquireC: make(chan bool, 1),
		timers:   make(map[string]*time.Timer),
		mu:       &sync.Mutex{},
		Stats:    &Metrics{CreatedAt: time.Now().Format(time.RFC3339)},
	}

	mutexes[Name] = semaphore

	// initial fill
	semaphore.acquireC <- true

	return semaphore
}

func getMutex(name string) (semaphore *Semaphore) {
	// a lock is just a semaphore with size=1
	mutexesMutex.RLock()
	mutex, ok := mutexes[name]
	mutexesMutex.RUnlock()

	if !ok {
		mutex = newSemaphore(name, 1)
	}

	return mutex
}
