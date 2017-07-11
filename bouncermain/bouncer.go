package bouncermain

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
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

	loglevel, err := logging.LogLevel(viper.GetString("logLevel"))
	if err != nil {
		panic("Invalid log level")
	}

	backend1Leveled.SetLevel(loglevel, "bouncer")

	logging.SetBackend(backend1Leveled)
}

func loadConfig() {
	viper.SetDefault("myHost", "0.0.0.0")
	viper.SetDefault("myPort", 5505)
	viper.SetDefault("logLevel", "INFO")
	viper.SetDefault("readTimeout", 30)
	viper.SetDefault("writeTimeout", 30)

	viper.BindEnv("myHost", "BOUNCER_HOST")
	viper.BindEnv("myPort", "BOUNCER_PORT")
	viper.BindEnv("logLevel", "BOUNCER_LOGLEVEL")
	viper.BindEnv("readTimeout", "BOUNCER_READ_TIMEOUT")
	viper.BindEnv("writeTimeout", "BOUNCER_WRITE_TIMEOUT")
}

func Main() {
	runtime.LockOSThread()
	loadConfig()
	setupLogging()

	addr := fmt.Sprintf("%v:%v", viper.GetString("myHost"), viper.GetInt("myPort"))

	logger.Info("Starting...")
	logger.Infof("Listening on %v", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      Router(),
		ReadTimeout:  time.Duration(viper.GetInt("readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("writeTimeout")) * time.Second,
	}

	logger.Fatal(server.ListenAndServe())

}
