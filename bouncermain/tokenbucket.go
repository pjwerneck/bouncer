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
	Size     uint64
	Interval time.Duration
	Stats    *Metrics
	acquireC chan bool
	timer    *time.Timer
}

var buckets = map[string]*TokenBucket{}
var bucketsMutex = &sync.Mutex{}

func newTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket) {
	bucket = &TokenBucket{
		Name:     name,
		Size:     size,
		Interval: interval,
		Stats:    &Metrics{CreatedAt: time.Now().Format(time.RFC3339)},
		acquireC: make(chan bool),
		timer:    time.NewTimer(interval),
	}

	buckets[name] = bucket

	go bucket.refill()
	//go bucket.Stats.Run()

	return bucket
}

func getTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket, err error) {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
	if !ok {
		bucket = newTokenBucket(name, size, interval)
		logger.Infof("tokenbucket created: name=%v, size=%v", name, size)
	}

	bucket.Size = size
	bucket.Interval = interval

	return
}

func (bucket *TokenBucket) refill() {
	var n uint64
	for {
		// bucket size being changed midloop is not a problem, since
		// we want size changes to take effect immediately
		for n = 0; n < bucket.Size; n++ {
			// make token available
			bucket.acquireC <- true
		}

		// wait
		<-bucket.timer.C

		// and reset the timer
		bucket.timer.Reset(bucket.Interval)
	}
}

func (bucket *TokenBucket) Acquire(maxwait time.Duration, arrival time.Time) (err error) {

	_, err = RecvTimeout(bucket.acquireC, maxwait)
	if err != nil {
		atomic.AddUint64(&bucket.Stats.TimedOut, 1)
		return err
	}

	wait := uint64(time.Since(arrival) / time.Millisecond)

	logger.Debugf("wait time %v", wait)

	atomic.AddUint64(&bucket.Stats.Acquired, 1)
	atomic.AddUint64(&bucket.Stats.WaitTime, wait)
	return nil
}

func (bucket *TokenBucket) GetStats() *Metrics {
	return bucket.Stats
}

func getTokenBucketStats(name string) (stats *Metrics, err error) {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
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
