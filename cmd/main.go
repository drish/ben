package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/drish/ben/config"
	"github.com/fatih/color"
)

var version = ""

func main() {
	trap()

	c, err := config.ReadConfig()
	if err != nil {
		output(err.Error())
		return
	}

	fmt.Println(c)
	// ben := ben.New()
	// ben.Run()
}

func output(s string) {
	color.Red(s)
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
