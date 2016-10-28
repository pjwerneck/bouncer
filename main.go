package main

import (
	"github.com/pjwerneck/bouncer/bouncermain"

	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	pprof.StopCPUProfile()

	log.Printf("Server stopped")
}
