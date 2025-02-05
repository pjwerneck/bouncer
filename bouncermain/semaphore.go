package bouncermain

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

type Semaphore struct {
	Name     string
	Size     uint64
	Keys     map[string]time.Duration
	acquireC chan bool
	timers   map[string]*time.Timer
	mu       *sync.Mutex
	Stats    *Metrics
}

var semaphores = map[string]*Semaphore{}
var semaphoresMutex = &sync.Mutex{}

func newSemaphore(name string, size uint64) (semaphore *Semaphore) {
	semaphore = &Semaphore{
		Name:     name,
		Size:     uint64(size),
		Keys:     make(map[string]time.Duration),
		acquireC: make(chan bool, size),
		timers:   make(map[string]*time.Timer),
		mu:       &sync.Mutex{},
		Stats:    &Metrics{CreatedAt: time.Now().Format(time.RFC3339)},
	}

	semaphores[name] = semaphore

	return semaphore
}

func getSemaphore(name string, size uint64) (semaphore *Semaphore, err error) {
	semaphoresMutex.Lock()
	defer semaphoresMutex.Unlock()

	semaphore, ok := semaphores[name]
	if !ok {
		semaphore = newSemaphore(name, size)
		logger.Infof("semaphore created: name=%v, size=%v", name, size)
	}

	semaphore.mu.Lock()
	semaphore.Size = size
	semaphore.mu.Unlock()

	return semaphore, err
}

func (semaphore *Semaphore) getKey(key string) (expires time.Duration, ok bool) {
	semaphore.mu.Lock()
	defer semaphore.mu.Unlock()
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
				logger.Debugf("semaphore expired: name=%v, key=%v", semaphore.Name, key)
				semaphore.delKey(key)
				atomic.AddUint64(&semaphore.Stats.Expired, 1)
			})
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
		logger.Debugf("semaphore reacquired: name=%v, key=%v", semaphore.Name, token)
		return token, nil
	}

	// otherwise, check for available slots
	started := time.Now()
	for {

		if semaphore.setKey(key, expires) {
			token = key
			atomic.AddUint64(&semaphore.Stats.Acquired, 1)
			break
		}

		if maxwait == 0 {
			atomic.AddUint64(&semaphore.Stats.TimedOut, 1)
			logger.Debugf("semaphore acquire timed out: name=%v, maxwait=%v", semaphore.Name, maxwait)
			return "", ErrTimedOut
		}

		if maxwait > 0 && time.Since(started) >= maxwait {
			atomic.AddUint64(&semaphore.Stats.TimedOut, 1)
			logger.Debugf("semaphore acquire timed out: name=%v, maxwait=%v", semaphore.Name, maxwait)
			return "", ErrTimedOut
		}

		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	logger.Debugf("semaphore acquired: name=%v, key=%v", semaphore.Name, token)
	return token, nil
}

func (semaphore *Semaphore) Release(key string) error {
	err := semaphore.delKey(key)
	atomic.AddUint64(&semaphore.Stats.Released, 1)
	logger.Debugf("semaphore released: name=%v, key=%v", semaphore.Name, key)
	return err
}

func (semaphore *Semaphore) GetStats() *Metrics {
	return semaphore.Stats
}

func getSemaphoreStats(name string) (stats *Metrics, err error) {
	semaphoresMutex.Lock()
	defer semaphoresMutex.Unlock()

	semaphore, ok := semaphores[name]
	if !ok {
		return nil, ErrNotFound
	}

	return semaphore.Stats, nil
}

// SemaphoreAcquireHandler godoc
// @Summary Acquire a semaphore
// @Description Acquire a semaphore lock.
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param size query int false "Semaphore size" default(1)
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Param expires query int false "Expiration time" default(60000)
// @Success 200 {string} Reply "The semaphore release key"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 408 {string} Reply "Request Timeout - `maxWait` exceeded"
// @Router /semaphore/{name}/acquire [get]
func SemaphoreAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.acquire: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		rep.Body, err = semaphore.Acquire(req.MaxWait, req.Expires, req.Key)
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

// SemaphoreReleaseHandler godoc
// @Summary Release a semaphore
// @Description Release a semaphore lock
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param size query int false "Semaphore size" default(1)
// @Param key query string true "Release key"
// @Success 204 "Semaphore released successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 409 {string} Reply "Conflict - key is invalid or already released"
// @Router /semaphore/{name}/release [get]
func SemaphoreReleaseHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.release: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		err = semaphore.Release(req.Key)
		rep.Status = http.StatusNoContent

		logger.Debugf("semaphore.keys: %+v", semaphore.Keys)
	}

	rep.WriteResponse(w, r, err)
}

// TODO: semaphore stats should have max_ever_held, currently_held, and total_held_time
func SemaphoreStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getSemaphoreStats)
}
