package bouncermain

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/julienschmidt/httprouter"
)

type TokenBucket struct {
	Name     string
	size     uint64
	tokens   int64
	interval time.Duration
	mu       *sync.RWMutex
	stopC    chan bool // Channel to stop the refill goroutine
	Stats    *Metrics
}

var tokenBuckets = map[string]*TokenBucket{}
var tokenBucketsMutex = &sync.Mutex{}

func newTokenBucket(name string, size uint64, interval time.Duration) *TokenBucket {
	bucket := &TokenBucket{
		Name:     name,
		size:     size,
		tokens:   int64(size),
		interval: interval,
		mu:       &sync.RWMutex{},
		stopC:    make(chan bool),
		Stats:    &Metrics{},
	}

	// Start refill goroutine
	go bucket.refill()

	tokenBuckets[name] = bucket
	return bucket
}

func getTokenBucket(name string, size uint64, interval time.Duration) (*TokenBucket, error) {
	tokenBucketsMutex.Lock()
	defer tokenBucketsMutex.Unlock()

	bucket, ok := tokenBuckets[name]
	if !ok {
		bucket = newTokenBucket(name, size, interval)
		logger.Infof("New token bucket created: name=%v, size=%v, interval=%v", name, size, interval)
		return bucket, nil
	}

	// If bucket exists but parameters changed, stop existing refill and start new one
	if bucket.size != size || bucket.interval != interval {
		bucket.mu.Lock()
		close(bucket.stopC) // Stop existing refill goroutine
		bucket.stopC = make(chan bool)
		bucket.size = size
		bucket.interval = interval
		atomic.StoreInt64(&bucket.tokens, int64(size))
		bucket.mu.Unlock()

		go bucket.refill()
	}

	return bucket, nil
}

func (bucket *TokenBucket) refill() {
	ticker := time.NewTicker(bucket.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bucket.stopC:
			return
		case <-ticker.C:
			bucket.mu.RLock()
			size := bucket.size
			bucket.mu.RUnlock()
			atomic.StoreInt64(&bucket.tokens, int64(size))
		}
	}
}

func (bucket *TokenBucket) Acquire(maxwait time.Duration, arrival time.Time) error {
	started := time.Now()

	for {
		tokens := atomic.LoadInt64(&bucket.tokens)
		if tokens > 0 && atomic.CompareAndSwapInt64(&bucket.tokens, tokens, tokens-1) {
			atomic.AddUint64(&bucket.Stats.Acquired, 1)
			return nil
		}

		if maxwait == 0 {
			atomic.AddUint64(&bucket.Stats.TimedOut, 1)
			return ErrTimedOut
		}

		if maxwait > 0 && time.Since(started) >= maxwait {
			atomic.AddUint64(&bucket.Stats.TimedOut, 1)
			return ErrTimedOut
		}

		time.Sleep(time.Millisecond)
	}
}

func (bucket *TokenBucket) GetStats() *Metrics {
	return bucket.Stats
}

func getTokenBucketStats(name string) (stats *Metrics, err error) {
	tokenBucketsMutex.Lock()
	defer tokenBucketsMutex.Unlock()

	bucket, ok := tokenBuckets[name]
	if !ok {
		return nil, ErrNotFound
	}

	return bucket.Stats, nil
}

// TokenBucketAcquireHandler godoc
// @Summary Acquire a token from a token bucket
// @Description Every `interval` milliseconds, the bucket is refilled with `size` tokens.
// @Description Each acquire request takes one token out of the bucket, or waits up to `maxWait` milliseconds for a token to be available.
// @Tags TokenBucket
// @Produce plain
// @Param name path string true "Token bucket name"
// @Param size query int false "Bucket size" default(1)
// @Param interval query int false "Refill interval" default(1000)
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Success 204 {string} Reply "Token acquired successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - token bucket not found
// @Failure 408 {string} Reply "Request Timeout - `maxWait` exceeded"
// @Router /tokenbucket/{name}/acquire [get]
func TokenBucketAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var bucket *TokenBucket

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("tokenbucket.acquire: %+v", req)
		bucket, err = getTokenBucket(ps[0].Value, req.Size, req.Interval)
	}

	if err == nil {
		err = bucket.Acquire(req.MaxWait, req.Arrival)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func TokenBucketStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getTokenBucketStats)
}
