package bouncermain_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

	// check
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
