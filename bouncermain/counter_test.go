package bouncermain_test

import (
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounterBasicOperations(t *testing.T) {
	baseURL := fmt.Sprintf("%s/counter/basic-test", server.URL)

	// Initial count should be 0
	status, body, err := GetRequest(fmt.Sprintf("%s/value", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "0", body)

	// Increment by 1 (default)
	status, body, err = GetRequest(fmt.Sprintf("%s/count", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "1", body)

	// Increment by 5
	status, body, err = GetRequest(fmt.Sprintf("%s/count?amount=5", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "6", body)

	// Decrement by 2
	status, body, err = GetRequest(fmt.Sprintf("%s/count?amount=-2", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "4", body)

	// Reset to 0
	status, _, err = GetRequest(fmt.Sprintf("%s/reset", baseURL))
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Verify reset
	status, body, err = GetRequest(fmt.Sprintf("%s/value", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "0", body)
}

func TestCounterConcurrentAccess(t *testing.T) {
	baseURL := fmt.Sprintf("%s/counter/concurrent-test", server.URL)
	var wg sync.WaitGroup
	results := make([]string, 100)

	// 100 concurrent increments
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, body, err := GetRequest(fmt.Sprintf("%s/count", baseURL))
			require.Nil(t, err)
			results[idx] = body
		}(i)
	}
	wg.Wait()

	// Convert results to integers and sort
	nums := make([]int, len(results))
	for i, v := range results {
		fmt.Sscanf(v, "%d", &nums[i])
	}
	sort.Ints(nums)

	// Verify we got all numbers 1-100 exactly once
	for i, v := range nums {
		require.Equal(t, i+1, v)
	}

	// Verify final value
	status, body, err := GetRequest(fmt.Sprintf("%s/value", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "100", body)
}

func TestCounterResetWithValue(t *testing.T) {
	baseURL := fmt.Sprintf("%s/counter/reset-test", server.URL)

	// Set initial value through increment
	status, _, err := GetRequest(fmt.Sprintf("%s/count?amount=42", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)

	// Reset to specific value
	status, _, err = GetRequest(fmt.Sprintf("%s/reset?value=10", baseURL))
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Verify value
	status, body, err := GetRequest(fmt.Sprintf("%s/value", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Equal(t, "10", body)
}

func TestCounterStats(t *testing.T) {
	baseURL := fmt.Sprintf("%s/counter/stats-test", server.URL)

	// Do some operations
	GetRequest(fmt.Sprintf("%s/count", baseURL))
	GetRequest(fmt.Sprintf("%s/count", baseURL))
	GetRequest(fmt.Sprintf("%s/reset", baseURL))
	GetRequest(fmt.Sprintf("%s/count", baseURL))

	// Check stats
	status, body, err := GetRequest(fmt.Sprintf("%s/stats", baseURL))
	require.Nil(t, err)
	require.Equal(t, 200, status)
	require.Contains(t, body, "acquired")
	require.Contains(t, body, "released")
}
