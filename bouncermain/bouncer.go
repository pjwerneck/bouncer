package bouncermain

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/spf13/viper"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var decoder = newDecoder()

func setupLogging() {
	level, err := zerolog.ParseLevel(viper.GetString("logLevel"))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}

func loadConfig() {
	viper.SetDefault("myHost", "0.0.0.0")
	viper.SetDefault("myPort", 5505)
	viper.SetDefault("logLevel", "INFO")
	viper.SetDefault("readTimeout", 30)
	viper.SetDefault("writeTimeout", 30)
	viper.SetDefault("maxSleepDuration", 5000) // 5 seconds in milliseconds

	viper.BindEnv("myHost", "BOUNCER_HOST")
	viper.BindEnv("myPort", "BOUNCER_PORT")
	viper.BindEnv("logLevel", "BOUNCER_LOGLEVEL")
	viper.BindEnv("readTimeout", "BOUNCER_READ_TIMEOUT")
	viper.BindEnv("writeTimeout", "BOUNCER_WRITE_TIMEOUT")
	viper.BindEnv("maxSleepDuration", "BOUNCER_MAX_SLEEP_DURATION")
}

var maxSleepDuration = time.Duration(viper.GetInt("maxSleepDuration")) * time.Millisecond

func Main() {
	runtime.LockOSThread()
	loadConfig()
	setupLogging()

	addr := fmt.Sprintf("%v:%v", viper.GetString("myHost"), viper.GetInt("myPort"))

	log.Info().Msg("Starting...")
	log.Info().Msgf("Listening on %v", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      Router(),
		ReadTimeout:  time.Duration(viper.GetInt("readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("writeTimeout")) * time.Second,
	}

	log.Fatal().Err(server.ListenAndServe())

}
