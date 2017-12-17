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

// Converts a string into a docker command
func PrepareCommand(command string) []string {
	return strings.Split(command, " ")
}
