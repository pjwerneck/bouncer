package main

import (
	"fmt"

	"github.com/pjwerneck/bouncer/bouncermain"
	_ "github.com/pjwerneck/bouncer/docs" // swagger docs

	"flag"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/rs/zerolog/log"
)

var (
	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file")
	versionFlag = flag.Bool("version", false, "print version information")
	version     = "dev"
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s\n",
			version)
		return
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal().Err(err).Msg("could not create CPU profile")
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
	}

	go bouncermain.Main()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	pprof.StopCPUProfile()

	log.Printf("Server stopped")
}
