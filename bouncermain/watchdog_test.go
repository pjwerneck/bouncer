package bouncermain_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWatchdogExpiration(t *testing.T) {
	// Create watchdog with 100ms expiration
	kickURL := fmt.Sprintf("%s/watchdog/expiration-test/kick?expires=100", server.URL)
	waitURL := fmt.Sprintf("%s/watchdog/expiration-test/wait?maxWait=200", server.URL)

	// Initial kick
	status, _, err := GetRequest(kickURL)
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Wait should succeed after expiration
	status, _, err = GetRequest(waitURL)
	require.Nil(t, err)
	require.Equal(t, 204, status)
}

func TestWatchdogKick(t *testing.T) {
	kickURL := fmt.Sprintf("%s/watchdog/kick-test/kick?expires=500", server.URL)
	waitURL := fmt.Sprintf("%s/watchdog/kick-test/wait?maxWait=100", server.URL)

	// Initial kick
	status, _, err := GetRequest(kickURL)
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Wait should timeout because watchdog was kicked
	status, _, err = GetRequest(waitURL)
	require.Nil(t, err)
	require.Equal(t, 408, status)
}

func TestWatchdogMultipleWaiters(t *testing.T) {
	kickURL := fmt.Sprintf("%s/watchdog/multi-test/kick?expires=100", server.URL)
	waitURL := fmt.Sprintf("%s/watchdog/multi-test/wait?maxWait=200", server.URL)

	// Initial kick
	status, _, err := GetRequest(kickURL)
	require.Nil(t, err)
	require.Equal(t, 204, status)

	// Launch multiple waiters
	var wg sync.WaitGroup
	results := make(chan int, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _, err := GetRequest(waitURL)
			require.Nil(t, err)
			results <- status
		}()
	}

	wg.Wait()
	close(results)

	// All waiters should succeed
	successCount := 0
	for status := range results {
		require.Equal(t, 204, status)
		successCount++
	}
	require.Equal(t, 5, successCount)
}
