package bouncermain_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEventWaitAndSend(t *testing.T) {
	time.AfterFunc(time.Duration(1e7),
		func() {
			status, _, err := GetRequest(fmt.Sprintf("%s/event/test1/send?message=completed", server.URL))
			require.Nil(t, err)
			require.Equal(t, 204, status)
		})

	status, body, err := GetRequest(fmt.Sprintf("%s/event/test1/wait?maxwait=100", server.URL))
	require.Nil(t, err)
	require.Equal(t, 200, status) // Changed from 204
	require.Equal(t, "completed", body)
}

func TestEventMultipleWaiters(t *testing.T) {
	baseURL := fmt.Sprintf("%s/event/multi-wait-test", server.URL)
	waitURL := fmt.Sprintf("%s/wait?maxwait=1000", baseURL)
	sendURL := fmt.Sprintf("%s/send?message=all-ready", baseURL)

	var wg sync.WaitGroup
	var successCount int32
	var correctMessage int32

	// Start multiple waiters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, body, err := GetRequest(waitURL)
			require.Nil(t, err)
			if status == 200 && body == "all-ready" {
				atomic.AddInt32(&successCount, 1)
				atomic.AddInt32(&correctMessage, 1)
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
	require.Equal(t, int32(5), correctMessage)
}
