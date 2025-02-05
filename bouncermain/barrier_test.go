package bouncermain_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBarrierTimeout(t *testing.T) {
	url := fmt.Sprintf("%s/barrier/timeout-test/wait?size=5&maxWait=100", server.URL)

	// Single request should timeout
	status, _, err := GetRequest(url)
	require.Nil(t, err)
	require.Equal(t, 408, status)
}

func TestBarrierSuccess(t *testing.T) {
	url := fmt.Sprintf("%s/barrier/success-test/wait?size=3", server.URL)
	var wg sync.WaitGroup
	results := make(chan int, 3)

	// Launch 3 concurrent requests
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _, err := GetRequest(url)
			require.Nil(t, err)
			results <- status
		}()
	}

	wg.Wait()
	close(results)

	// All requests should succeed
	successCount := 0
	for status := range results {
		require.Equal(t, 204, status)
		successCount++
	}
	require.Equal(t, 3, successCount)
}

func TestBarrierReuse(t *testing.T) {
	url := fmt.Sprintf("%s/barrier/reuse-test/wait?size=5", server.URL)

	// First round
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _, err := GetRequest(url)
			require.Nil(t, err)
			require.Equal(t, 204, status)
		}()
	}
	wg.Wait()

	// Second round - barrier should be reusable
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _, err := GetRequest(url)
			require.Nil(t, err)
			require.Equal(t, 204, status)
		}()
	}
	wg.Wait()
}
