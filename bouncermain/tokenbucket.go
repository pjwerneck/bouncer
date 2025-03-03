package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type TokenBucketStats struct {
	Acquired        uint64  `json:"acquired"`
	TotalWaitTime   uint64  `json:"total_wait_time"`
	TimedOut        uint64  `json:"timed_out"`
	CreatedAt       string  `json:"created_at"`
	AverageWaitTime float64 `json:"average_wait_time"`
}

type TokenBucket struct {
	Name     string
	Size     uint64        // private field
	interval time.Duration // private field
	Stats    *TokenBucketStats
	mu       sync.RWMutex // protect size and interval

	available  int64 // atomic counter for available tokens
	nextRefill int64 // atomic unix nano for next refill
}

var buckets = map[string]*TokenBucket{}
var bucketsMutex = &sync.RWMutex{}

func newTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket) {
	now := time.Now()
	bucket = &TokenBucket{
		Name:       name,
		Size:       size,
		interval:   interval,
		Stats:      &TokenBucketStats{CreatedAt: now.Format(time.RFC3339)},
		available:  int64(size),
		nextRefill: now.Add(interval).UnixNano(),
	}

	buckets[name] = bucket
	return bucket
}

func getTokenBucket(name string, size uint64, interval time.Duration) (bucket *TokenBucket, err error) {
	bucketsMutex.RLock()
	bucket, ok := buckets[name]
	bucketsMutex.RUnlock()

	if ok {
		bucket.mu.RLock()
		currentSize := bucket.Size
		currentInterval := bucket.interval
		bucket.mu.RUnlock()

		if size != currentSize {
			log.Warn().
				Str("name", name).
				Uint64("current_size", currentSize).
				Uint64("new_size", size).
				Msg("tokenbucket size modification through acquire is deprecated and will be removed in a future version")
		}

		if size != currentSize || interval != currentInterval {
			bucket.mu.Lock()
			bucket.Size = size
			bucket.interval = interval
			bucket.mu.Unlock()
		}
		return bucket, nil
	}

	// Bucket doesn't exist, need to create it
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	// Check again in case another goroutine created it
	bucket, ok = buckets[name]
	if ok {
		return bucket, nil
	}

	bucket = newTokenBucket(name, size, interval)
	return bucket, nil
}

func (bucket *TokenBucket) refillTokens() {
	now := time.Now().UnixNano()
	next := atomic.LoadInt64(&bucket.nextRefill)

	// Check if refill is due
	if now < next {
		return
	}

	// Try to update nextRefill - if we fail, someone else already did it
	if atomic.CompareAndSwapInt64(&bucket.nextRefill, next, now+bucket.interval.Nanoseconds()) {
		bucket.mu.RLock()
		size := bucket.Size
		bucket.mu.RUnlock()
		atomic.StoreInt64(&bucket.available, int64(size))
	}
}

func (bucket *TokenBucket) Acquire(maxwait time.Duration, arrival time.Time) error {
	deadline := time.Now().Add(maxwait)

	for {
		bucket.refillTokens()

		// Try to acquire a token
		for {
			current := atomic.LoadInt64(&bucket.available)
			if current <= 0 {
				break
			}
			if atomic.CompareAndSwapInt64(&bucket.available, current, current-1) {
				wait := uint64(time.Since(arrival) / time.Millisecond)
				atomic.AddUint64(&bucket.Stats.Acquired, 1)
				atomic.AddUint64(&bucket.Stats.TotalWaitTime, wait)
				return nil
			}
		}

		// No tokens available, check timeout
		if maxwait >= 0 && time.Now().After(deadline) {
			atomic.AddUint64(&bucket.Stats.TimedOut, 1)
			return ErrTimedOut
		}

		// Sleep until next refill or deadline
		now := time.Now()
		sleepUntil := time.Unix(0, atomic.LoadInt64(&bucket.nextRefill))
		if maxwait >= 0 && deadline.Before(sleepUntil) {
			sleepUntil = deadline
		}

		time.Sleep(min(sleepUntil.Sub(now), maxSleepDuration))
	}
}

func deleteTokenBucket(name string) error {
	bucketsMutex.Lock()
	defer bucketsMutex.Unlock()

	_, ok := buckets[name]
	if !ok {
		return ErrNotFound
	}

	delete(buckets, name) // Remove from global map
	return nil
}

func getTokenBucketStats(name string) (interface{}, error) {
	bucketsMutex.RLock()
	bucket, ok := buckets[name]
	bucketsMutex.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}

	stats := &TokenBucketStats{}
	*stats = *bucket.Stats // Copy stats
	acquired := atomic.LoadUint64(&bucket.Stats.Acquired)
	if acquired > 0 {
		stats.AverageWaitTime = float64(atomic.LoadUint64(&bucket.Stats.TotalWaitTime)) / float64(acquired)
	}

	return stats, nil
}
