package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
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
		acquireC: make(chan bool, 1),
		timer:    time.NewTimer(interval),
	}

	buckets[name] = bucket

	go bucket.refill()
	//go bucket.Stats.Run()

	return bucket
}

func getTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket) {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
	if !ok {
		bucket = newTokenBucket(name, size, interval)
		logger.Infof("New token bucket: name=%v, size=%v", name, size)
	}

	bucket.Size = size
	bucket.Interval = interval

	return bucket
}

func (bucket *TokenBucket) refill() {
	for {
		size := bucket.Size

		for size > 0 {
			// make token available
			bucket.acquireC <- true

			size--
		}

		// wait
		<-bucket.timer.C
		logger.Debugf("tokenbucket %v refill", bucket.Name)

		// and reset the timer
		bucket.timer.Reset(bucket.Interval)
	}
}

func (bucket *TokenBucket) Acquire(timeout time.Duration) (token string, err error) {

	acquired, err := RecvTimeout(bucket.acquireC, timeout)
	if err != nil {
		atomic.AddUint64(&bucket.Stats.TimedOut, 1)
		return token, err
	}

	// if token is valid, just return a valid token
	if acquired {
		atomic.AddUint64(&bucket.Stats.Acquired, 1)
		token = "OK"
		return token, nil
	} else {
		// not supposed to be here, as all values in acquireC are true
		panic("we're not supposed to be here")
	}

	return token, nil
}

func (bucket *TokenBucket) GetStats() *Metrics {
	return bucket.Stats
}
