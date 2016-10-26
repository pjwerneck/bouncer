package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type TokenBucket struct {
	Name     string
	Size     uint64
	Tokens   int64
	acquireC chan bool
	timer    *time.Timer
	Stats    *Metrics
}

var buckets = map[string]*TokenBucket{}
var bucketsMutex = &sync.RWMutex{}

func newTokenBucket(name string, size uint64) (bucket *TokenBucket) {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
	if ok {
		return bucket
	}

	bucket = &TokenBucket{
		Name:     name,
		Size:     size,
		Tokens:   int64(size), // an uint64 overflows when decrement below zero
		acquireC: make(chan bool, 1),
		timer:    time.NewTimer(time.Second),
		Stats:    &Metrics{CreatedAt: time.Now().Format(time.RFC3339)},
	}

	buckets[name] = bucket

	go bucket.refill()
	//go bucket.Stats.Run()

	return bucket
}

func getTokenBucket(name string, size uint64) (bucket *TokenBucket) {
	// most of the time we'll hold the R lock for just a sec
	bucketsMutex.RLock()
	bucket, ok := buckets[name]
	bucketsMutex.RUnlock()

	if !ok {
		logger.Infof("New token bucket: name=%v, size=%v", name, size)
		bucket = newTokenBucket(name, size)
	}

	bucket.setSize(size)

	return bucket
}

func (bucket *TokenBucket) setSize(size uint64) {
	atomic.StoreUint64(&bucket.Size, size)
	atomic.StoreUint64(&bucket.Stats.Nominal, size)
}

func (bucket *TokenBucket) getSize() uint {
	return uint(atomic.LoadUint64(&bucket.Size))
}

func (bucket *TokenBucket) refill() {
	for {
		// decrement usage by one
		tokens := atomic.AddInt64(&bucket.Tokens, ^int64(0))

		if tokens >= 0 {
			// if ok, make token available and continue
			bucket.acquireC <- true

		} else {
			// otherwise, wait and reset
			<-bucket.timer.C
			// and refill
			atomic.StoreInt64(&bucket.Tokens, int64(atomic.LoadUint64(&bucket.Size)))
			bucket.timer.Reset(time.Second)
		}
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
		// not supposed to be here, as all values in recv are true
		panic("we're not supposed to be here")
	}

	return token, nil
}

func (bucket *TokenBucket) GetStats() *Metrics {
	return bucket.Stats
}
