package bouncermain_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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

func DeleteRequest(url string) (status int, body string, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	rep, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer rep.Body.Close() // Add this line to properly close the response body

	bs, err := io.ReadAll(rep.Body)
	if err != nil {
		return
	}

	body = string(bs)
	status = rep.StatusCode

	return
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
