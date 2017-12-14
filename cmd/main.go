package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/drish/ben"
	"github.com/drish/ben/config"
	"github.com/drish/ben/utils"
)

var version = ""

func main() {
	trap()

	c, err := config.ReadConfig("ben.json")
	if err != nil {
		utils.Fatal(err)
	}

	err = ben.New(c).Run()

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
		println("\n")
		os.Exit(1)
	}()
}
