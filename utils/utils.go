package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func Contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "\n     %s %s\n\n", color.RedString("Error:"), err)
	os.Exit(1)
}

func PrepareImage(name, version string) string {
	return name + ":" + version
}

// PrepareCommand prepares the benchmark command
// if empty returns default command
func PrepareCommand(command string) []string {
	if command == "" {
		return []string{"go", "test", "-bench=."}
	}
	return strings.Split(command, " ")
}
