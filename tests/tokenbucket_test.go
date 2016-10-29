package bouncermain_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pjwerneck/bouncer/bouncermain"
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

func TestGetToken(t *testing.T) {
	url := fmt.Sprintf("%s/v1/tokenbucket/testingbucket/acquire?size=10", server.URL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}

	rep, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	if rep.StatusCode != 204 {
		t.Errorf("204 No Content expected: %d", rep.StatusCode)
	}

}

func TestGetTokens(t *testing.T) {
	url := fmt.Sprintf("%s/v1/tokenbucket/testingbucket/acquire?size=10&maxwait=1", server.URL)

	for i := 0; i < 9; i++ {
		status, body, err := GetRequest(url)
		if err != nil {
			t.Error(err)
		}
		if status != 204 {
			t.Errorf("204 No Content expected: %v %v", status, body)
		}
	}

	status, body, err := GetRequest(url)
	if err != nil {
		t.Error(err)
	}
	if status != 408 {
		t.Errorf("408 Request Timeout expected: %v %v", status, body)
	}

}
