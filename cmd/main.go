package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/drish/ben"
	"github.com/drish/ben/config"
	"github.com/drish/ben/utils"
)

var usage = `Usage: ben [options...]
Options:
  -o  output file. Default is ./benchmarks.md
  -d  display benchmark results to stdout. Default is false.
`

var defaultBenchmarkFile = "./ben-summary.html"

// TODO: add before environment
func main() {
	trap()

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage))
	}

	outputFlag := flag.String("o", defaultBenchmarkFile, "output summary file")
	displayFlag := flag.Bool("d", false, "display benchmark results to stdout")
	flag.Parse()

	c, err := config.ReadConfig("ben.json")
	if err != nil {
		utils.Fatal(err)
	}

	err = ben.New(c).Run(*outputFlag, *displayFlag)
	if err != nil {
		utils.Fatal(err)
	}
}

// trappy
func trap() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<-sigs
		fmt.Println()
		os.Exit(1)
	}()
}
