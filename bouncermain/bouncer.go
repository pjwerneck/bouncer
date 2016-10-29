package bouncermain

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("bouncer")

var decoder = newDecoder()

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

func setupLogging() {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)

	logging.SetBackend(backend1Leveled)
}

func Main() {
	runtime.LockOSThread()
	setupLogging()

	logger.Info("Starting...")

	server := &http.Server{
		Addr:         ":5505",
		Handler:      Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Fatal(server.ListenAndServe())

}
