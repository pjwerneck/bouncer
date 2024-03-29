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
