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

const (
	Version = "0.2.0"
)

var usage = `Usage: ben [options...]
Options:
  -o  output file. Default is ./benchmarks.md
  -d  display benchmark results to stdout. Default is false.
  -v  prints current version
`

var defaultBenchmarkFile = "./benchmarks.md"

func main() {
	trap()

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage))
	}

	outputFlag := flag.String("o", defaultBenchmarkFile, "OPTIONAL output summary file")
	displayFlag := flag.Bool("d", false, "OPTIONAL display benchmark results to stdout")
	vFlag := flag.Bool("v", false, "prints current version")
	flag.Parse()

	if *vFlag {
		fmt.Printf("\n\r  Ben version %s\n\n", Version)
		os.Exit(0)
	}

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
