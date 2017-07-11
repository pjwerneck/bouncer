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
	primary   *httptest.Server
	secondary *httptest.Server
)

func init() {
	primary = httptest.NewServer(bouncermain.Router())
	secondary = httptest.NewServer(bouncermain.Router())
}
