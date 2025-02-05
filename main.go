package main

import (
	"github.com/pjwerneck/bouncer/bouncermain"
	_ "github.com/pjwerneck/bouncer/docs" // swagger docs

	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

// @title Bouncer API
// @version 0.1.6
// @description Bouncer is a simple HTTP server that provides rate limiting and synchronization primitives for distributed systems.

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
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
