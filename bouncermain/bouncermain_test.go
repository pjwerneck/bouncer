package bouncermain_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pjwerneck/bouncer/bouncermain"
	"github.com/stretchr/testify/require"
)

var (
	server *httptest.Server
	reader io.Reader
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

	bs, err := ioutil.ReadAll(rep.Body)
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
	url := fmt.Sprintf("%s/tokenbucket/test2/acquire?size=10&maxwait=1&interval=10", server.URL)

	n := 0
	for {
		status, _, err := GetRequest(url)
		require.Nil(t, err)
		if status != 204 {
			break
		}
		n++
	}
	require.Equal(t, 10, n)

	time.Sleep(time.Duration(10 * 1e6))
	for {
		status, _, err := GetRequest(url)
		require.Nil(t, err)
		if status != 204 {
			break
		}
		n++
	}
	require.Equal(t, 20, n)

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
	status, key, err = GetRequest(url)
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
