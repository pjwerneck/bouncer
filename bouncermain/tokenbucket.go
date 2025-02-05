package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type TokenBucketStats struct {
	Acquired      uint64  `json:"acquired"`
	TotalWaitTime uint64  `json:"total_wait_time"`
	AvgWaitTime   float64 `json:"average_wait_time"`
	TimedOut      uint64  `json:"timed_out"`
	CreatedAt     string  `json:"created_at"`
}

type TokenBucket struct {
	Name     string
	size     uint64        // private field
	interval time.Duration // private field
	Stats    *TokenBucketStats
	acquireC chan bool
	timer    *time.Timer
	mu       sync.RWMutex // protect size and interval
}

var buckets = map[string]*TokenBucket{}
var bucketsMutex = &sync.Mutex{}

func newTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket) {
	bucket = &TokenBucket{
		Name:     name,
		size:     size,
		interval: interval,
		Stats:    &TokenBucketStats{CreatedAt: time.Now().Format(time.RFC3339)},
		acquireC: make(chan bool),
		timer:    time.NewTimer(interval),
	}

	buckets[name] = bucket

	go bucket.refill()

	return bucket
}

func getTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket, err error) {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
	if !ok {
		bucket = newTokenBucket(name, size, interval)
		logger.Infof("tokenbucket created: name=%v, size=%v", name, size)
		return
	}

	// Check current values before acquiring lock
	bucket.mu.RLock()
	currentSize := bucket.size
	currentInterval := bucket.interval
	bucket.mu.RUnlock()

	// Only update if values actually changed
	if size != currentSize || interval != currentInterval {
		bucket.mu.Lock()
		bucket.size = size
		bucket.interval = interval
		bucket.mu.Unlock()
	}

	return
}

func (bucket *TokenBucket) refill() {
	for {
		bucket.mu.RLock()
		size := bucket.size
		interval := bucket.interval
		bucket.mu.RUnlock()

		for n := uint64(0); n < size; n++ {
			bucket.acquireC <- true
		}

		<-bucket.timer.C
		bucket.timer.Reset(interval)
	}
}

// Add these getter methods for size and interval
func (bucket *TokenBucket) Size() uint64 {
	bucket.mu.RLock()
	defer bucket.mu.RUnlock()
	return bucket.size
}

func (bucket *TokenBucket) Interval() time.Duration {
	bucket.mu.RLock()
	defer bucket.mu.RUnlock()
	return bucket.interval
}

func (bucket *TokenBucket) Acquire(maxwait time.Duration, arrival time.Time) (err error) {

	_, err = RecvTimeout(bucket.acquireC, maxwait)
	if err != nil {
		atomic.AddUint64(&bucket.Stats.TimedOut, 1)
		return err
	}

	wait := uint64(time.Since(arrival) / time.Millisecond)

	atomic.AddUint64(&bucket.Stats.Acquired, 1)
	atomic.AddUint64(&bucket.Stats.TotalWaitTime, wait)

	// Update average wait time
	acquired := atomic.LoadUint64(&bucket.Stats.Acquired)
	totalWait := atomic.LoadUint64(&bucket.Stats.TotalWaitTime)
	if acquired > 0 {
		bucket.Stats.AvgWaitTime = float64(totalWait) / float64(acquired)
	}

	return nil
}

func (bucket *TokenBucket) GetStats() *TokenBucketStats {
	return bucket.Stats
}

func getTokenBucketStats(name string) (stats *TokenBucketStats, err error) {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
	if !ok {
		return nil, ErrNotFound
	}

	return bucket.Stats, nil
}

func deleteTokenBucket(name string) error {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	bucket, ok := buckets[name]
	if !ok {
		return ErrNotFound
	}

	// Stop the refill goroutine by closing acquireC
	close(bucket.acquireC)
	delete(buckets, name)
	return nil
}
