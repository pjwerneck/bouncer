package bouncermain_test

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/pjwerneck/bouncer/bouncermain"
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
