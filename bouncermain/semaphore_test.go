package bouncermain_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSemaphoreAcquireAndRelease(t *testing.T) {
	url := fmt.Sprintf("%s/semaphore/test1/acquire?maxWait=1", server.URL)

	status, key, err := GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 200, status)

	status, _, err = GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 408, status)

	url = fmt.Sprintf("%s/semaphore/test1/release?key=%s", server.URL, key)
	status, _, err = GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 204, status)

}

func TestSemaphoreConcurrentAccess(t *testing.T) {
	baseURL := fmt.Sprintf("%s/semaphore/concurrent-test", server.URL)
	acquireURL := fmt.Sprintf("%s/acquire?size=2&maxwait=100", baseURL)

	// Test concurrent access
	var keys []string
	var keysMutex sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, key, _ := GetRequest(acquireURL)
			if status == 200 {
				keysMutex.Lock()
				keys = append(keys, key)
				keysMutex.Unlock()
			}
		}()
	}
	wg.Wait()

	// Only 2 requests should succeed
	keysMutex.Lock()
	keysLen := len(keys)
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	keysMutex.Unlock()

	require.Equal(t, 2, keysLen)

	// Release the semaphores
	for _, key := range keysCopy {
		releaseURL := fmt.Sprintf("%s/release?key=%s", baseURL, key)
		status, _, err := GetRequest(releaseURL)
		require.Nil(t, err)
		require.Equal(t, 204, status)
	}
}

func TestSemaphoreExpiration(t *testing.T) {
	baseURL := fmt.Sprintf("%s/semaphore/expiration-test", server.URL)
	acquireURL := fmt.Sprintf("%s/acquire?expires=100", baseURL)

	// Initial acquire
	status, key, err := GetRequest(acquireURL)
	require.Nil(t, err)
	require.Equal(t, 200, status)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be able to acquire again
	status, _, err = GetRequest(acquireURL)
	require.Nil(t, err)
	require.Equal(t, 200, status)

	// Old key should fail to release
	releaseURL := fmt.Sprintf("%s/release?key=%s", baseURL, key)
	status, _, err = GetRequest(releaseURL)
	require.Nil(t, err)
	require.Equal(t, 409, status)

	// Verify stats
	status, body, err := GetRequest(fmt.Sprintf("%s/stats", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Contains(t, body, `"expired":1`)
}

func TestSemaphoreStats(t *testing.T) {
	baseURL := fmt.Sprintf("%s/semaphore/stats-test", server.URL)
	acquireURL := fmt.Sprintf("%s/acquire?size=2&maxwait=100", baseURL)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			GetRequest(acquireURL)
		}()
	}
	wg.Wait()

	status, body, err := GetRequest(fmt.Sprintf("%s/stats", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Contains(t, body, `"acquired":2`)
	require.Contains(t, body, `"timed_out":1`)
	require.Contains(t, body, `"max_ever_held":2`)
}

func TestSemaphoreNegativeSize(t *testing.T) {
	baseURL := fmt.Sprintf("%s/semaphore/negative-size-test", server.URL)
	acquireURL := fmt.Sprintf("%s/acquire?size=-1&maxwait=1", baseURL)

	status, _, err := GetRequest(acquireURL)
	require.Nil(t, err)
	require.Equal(t, 400, status)

	acquireURL = fmt.Sprintf("%s/acquire?size=-42&maxwait=1", baseURL)
	status, _, err = GetRequest(acquireURL)
	require.Nil(t, err)
	require.Equal(t, 400, status)
}
