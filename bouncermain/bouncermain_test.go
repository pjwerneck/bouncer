package bouncermain_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pjwerneck/bouncer/bouncermain"
	"github.com/stretchr/testify/require"
)

var (
	server *httptest.Server
)

func init() {
	server = httptest.NewServer(bouncermain.Router())
}

func GetRequest(url string) (status int, body string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	rep, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	bs, err := io.ReadAll(rep.Body)
	if err != nil {
		return
	}

	body = string(bs)
	status = rep.StatusCode

	return
}

func TestGetTokensUntilEmpty(t *testing.T) {
	url := fmt.Sprintf("%s/tokenbucket/test1/acquire?size=100&maxwait=1", server.URL)
	n := 0
	for {
		status, _, err := GetRequest(url)
		require.Nil(t, err)
		if status != 204 {
			break
		}
		n++
	}
	require.Equal(t, 100, n)
}

func TestGetTokensUntilEmptyAndWaitForRefill(t *testing.T) {
	// Use larger intervals to make the test more reliable
	url := fmt.Sprintf("%s/tokenbucket/test2/acquire?size=10&maxwait=10&interval=1000", server.URL)
	var count int32
	var phase2count int32

	// First phase: drain the bucket
	for {
		status, _, err := GetRequest(url)
		require.Nil(t, err)
		if status != 204 {
			require.Equal(t, 408, status) // Ensure it's a timeout
			break
		}
		atomic.AddInt32(&count, 1)
	}

	firstCount := atomic.LoadInt32(&count)
	require.Equal(t, int32(10), firstCount)

	// Wait for at least one refill cycle
	time.Sleep(1100 * time.Millisecond)

	// Second phase: try to get exactly 10 more tokens
	for i := 0; i < 10; i++ {
		status, _, err := GetRequest(url)
		require.Nil(t, err)
		if status == 204 {
			atomic.AddInt32(&phase2count, 1)
		}
	}

	// Verify we got exactly 10 more tokens
	require.Equal(t, int32(10), atomic.LoadInt32(&phase2count))
	require.Equal(t, int32(20), atomic.LoadInt32(&count)+atomic.LoadInt32(&phase2count))
}

func TestSemaphoreAcquireAndRelease(t *testing.T) {
	url := fmt.Sprintf("%s/semaphore/test1/acquire?maxwait=1", server.URL)

	status, key, err := GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 200, status)

	status, _, err = GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 408, status)

	url = fmt.Sprintf("%s/semaphore/test1/release?maxwait=1&key=%s", server.URL, key)
	status, _, err = GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 204, status)

}

func TestEventWaitAndSend(t *testing.T) {
	time.AfterFunc(time.Duration(1e7),
		func() {

			status, _, err := GetRequest(fmt.Sprintf("%s/event/test1/send", server.URL))
			require.Nil(t, err)
			require.Equal(t, 204, status)
		})

	status, _, err := GetRequest(fmt.Sprintf("%s/event/test1/wait?maxwait=100", server.URL))
	require.Nil(t, err)
	require.Equal(t, 204, status)
}

func TestTokenBucketTimeout(t *testing.T) {
	url := fmt.Sprintf("%s/tokenbucket/timeout-test/acquire?size=1&maxwait=100", server.URL)

	// First request should succeed
	status, _, err := GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Second request should timeout
	status, _, err = GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 408, status)
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

func TestEventMultipleWaiters(t *testing.T) {
	baseURL := fmt.Sprintf("%s/event/multi-wait-test", server.URL)
	waitURL := fmt.Sprintf("%s/wait?maxwait=1000", baseURL)
	sendURL := fmt.Sprintf("%s/send", baseURL)

	var wg sync.WaitGroup
	var successCount int32 // Changed from int to int32

	// Start multiple waiters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _, err := GetRequest(waitURL)
			require.Nil(t, err)
			if status == 204 {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	// Wait a bit and send the event
	time.Sleep(100 * time.Millisecond)
	status, _, err := GetRequest(sendURL)
	require.Nil(t, err)
	require.Equal(t, 204, status)

	wg.Wait()
	require.Equal(t, int32(5), successCount)
}

func TestStatsEndpoints(t *testing.T) {
	// Create and use a token bucket
	tbURL := fmt.Sprintf("%s/tokenbucket/stats-test/acquire?size=10", server.URL)
	status, _, err := GetRequest(tbURL)
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Check stats
	statsURL := fmt.Sprintf("%s/tokenbucket/stats-test/stats", server.URL)
	status, body, err := GetRequest(statsURL)
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Contains(t, body, "acquired")
	require.Contains(t, body, "total_wait_time")
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "Invalid token bucket size",
			url:      fmt.Sprintf("%s/tokenbucket/error-test/acquire?size=abc", server.URL),
			expected: 400,
		},
		{
			name:     "Invalid semaphore key",
			url:      fmt.Sprintf("%s/semaphore/error-test/release?key=invalid", server.URL),
			expected: 409,
		},
		{
			name:     "Non-existent stats",
			url:      fmt.Sprintf("%s/tokenbucket/nonexistent/stats", server.URL),
			expected: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, err := GetRequest(tt.url)
			require.Nil(t, err)
			require.Equal(t, tt.expected, status)
		})
	}
}
