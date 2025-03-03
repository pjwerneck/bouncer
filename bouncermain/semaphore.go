package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"
)

type SemaphoreStats struct {
	Acquired        uint64  `json:"acquired"`
	Reacquired      uint64  `json:"reacquired"`
	Released        uint64  `json:"released"`
	Expired         uint64  `json:"expired"`
	TotalWaitTime   uint64  `json:"total_wait_time"`
	AverageWaitTime float64 `json:"average_wait_time"`
	TimedOut        uint64  `json:"timed_out"`
	MaxEverHeld     uint64  `json:"max_ever_held"`
	CreatedAt       string  `json:"created_at"`
}

type Semaphore struct {
	Name     string
	Size     uint64
	Keys     map[string]time.Duration
	acquireC chan bool
	timers   map[string]*time.Timer
	mu       *sync.RWMutex
	Stats    *SemaphoreStats
}

var semaphores = map[string]*Semaphore{}
var semaphoresMutex = &sync.RWMutex{}

func newSemaphore(name string, size uint64) (semaphore *Semaphore) {
	semaphore = &Semaphore{
		Name:     name,
		Size:     uint64(size),
		Keys:     make(map[string]time.Duration),
		acquireC: make(chan bool, size),
		timers:   make(map[string]*time.Timer),
		mu:       &sync.RWMutex{},
		Stats:    &SemaphoreStats{CreatedAt: time.Now().Format(time.RFC3339)},
	}

	semaphores[name] = semaphore

	return semaphore
}

func getSemaphore(name string, size uint64) (semaphore *Semaphore, err error) {
	semaphoresMutex.RLock()
	semaphore, ok := semaphores[name]
	semaphoresMutex.RUnlock()

	if ok {
		// Check if we need to update settings
		semaphore.mu.RLock()
		currentSize := semaphore.Size
		semaphore.mu.RUnlock()

		if size != currentSize {
			log.Warn().
				Str("name", name).
				Uint64("current_size", currentSize).
				Uint64("new_size", size).
				Msg("semaphore size modification through acquire is deprecated and will be removed in a future version")

			semaphore.mu.Lock()
			semaphore.Size = size
			semaphore.mu.Unlock()
		}

		return semaphore, nil
	}

	// Semaphore doesn't exist, need to create it
	semaphoresMutex.Lock()
	defer semaphoresMutex.Unlock()

	// Check again in case another goroutine created it
	semaphore, ok = semaphores[name]
	if ok {
		return semaphore, nil
	}

	semaphore = newSemaphore(name, size)
	return semaphore, nil
}

func (semaphore *Semaphore) getKey(key string) (expires time.Duration, ok bool) {
	semaphore.mu.RLock()
	defer semaphore.mu.RUnlock()
	expires, ok = semaphore.Keys[key]
	return
}

func (semaphore *Semaphore) setKey(key string, expires time.Duration) bool {
	semaphore.mu.Lock()
	defer semaphore.mu.Unlock()

	if int(semaphore.Size)-len(semaphore.Keys) <= 0 {
		return false
	}

	semaphore.Keys[key] = expires
	if expires > 0 {
		semaphore.timers[key] = time.AfterFunc(expires,
			func() {
				log.Debug().Msgf("semaphore expired: name=%v, key=%v", semaphore.Name, key)
				semaphore.delKey(key)
				atomic.AddUint64(&semaphore.Stats.Expired, 1)
			})
	}

	// Update max ever held while holding the mutex
	current := uint64(len(semaphore.Keys))
	for {
		max := atomic.LoadUint64(&semaphore.Stats.MaxEverHeld)
		if current <= max || atomic.CompareAndSwapUint64(&semaphore.Stats.MaxEverHeld, max, current) {
			break
		}
	}

	return true
}

func (semaphore *Semaphore) delKey(key string) error {
	semaphore.mu.Lock()
	defer semaphore.mu.Unlock()

	if t, ok := semaphore.timers[key]; ok {
		t.Stop()
		delete(semaphore.timers, key)
	}

	if _, ok := semaphore.Keys[key]; ok {
		delete(semaphore.Keys, key)
	} else {
		return ErrKeyError
	}
	return nil
}

func (semaphore *Semaphore) Acquire(maxwait time.Duration, expires time.Duration, key string) (token string, err error) {
	// generate a random uuid as key if not provided
	if key == "" {
		key = uuid.Must(uuid.NewV4()).String()
	}

	// if there's an active token with this key, reacquire and return immediately
	_, ok := semaphore.getKey(key)

	if ok {
		token = key
		atomic.AddUint64(&semaphore.Stats.Reacquired, 1)
		return token, nil
	}

	// otherwise, check for available slots
	started := time.Now()
	for {

		if semaphore.setKey(key, expires) {
			token = key
			atomic.AddUint64(&semaphore.Stats.Acquired, 1)
			wait := uint64(time.Since(started) / time.Millisecond)
			atomic.AddUint64(&semaphore.Stats.TotalWaitTime, wait)

			break
		}

		if maxwait == 0 {
			atomic.AddUint64(&semaphore.Stats.TimedOut, 1)
			log.Debug().Msgf("semaphore acquire timed out: name=%v, maxwait=%v", semaphore.Name, maxwait)
			return "", ErrTimedOut
		}

		if maxwait > 0 && time.Since(started) >= maxwait {
			atomic.AddUint64(&semaphore.Stats.TimedOut, 1)
			log.Debug().Msgf("semaphore acquire timed out: name=%v, maxwait=%v", semaphore.Name, maxwait)
			return "", ErrTimedOut
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return token, nil
}

func (semaphore *Semaphore) Release(key string) error {
	err := semaphore.delKey(key)
	atomic.AddUint64(&semaphore.Stats.Released, 1)
	return err
}

func (semaphore *Semaphore) GetStats() *SemaphoreStats {
	return semaphore.Stats
}

func getSemaphoreStats(name string) (interface{}, error) {
	semaphoresMutex.RLock()
	defer semaphoresMutex.RUnlock()

	semaphore, ok := semaphores[name]
	if !ok {
		return nil, ErrNotFound
	}

	stats := &SemaphoreStats{}
	*stats = *semaphore.Stats
	totalWait := atomic.LoadUint64(&semaphore.Stats.TotalWaitTime)
	acquired := atomic.LoadUint64(&semaphore.Stats.Acquired)
	if acquired > 0 {
		stats.AverageWaitTime = float64(float64(totalWait) / float64(acquired))
	}
	return stats, nil
}

func deleteSemaphore(name string) error {
	semaphoresMutex.Lock()
	defer semaphoresMutex.Unlock()

	semaphore, ok := semaphores[name]
	if !ok {
		return ErrNotFound
	}

	semaphore.mu.Lock()
	defer semaphore.mu.Unlock()

	// Stop all timers
	for _, timer := range semaphore.timers {
		timer.Stop()
	}

	delete(semaphores, name)
	return nil
}
