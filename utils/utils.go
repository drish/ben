package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func StripCtlAndExtFromUnicode(str string) string {
	isOk := func(r rune) bool {
		// skip newline char
		if r == 10 {
			return false
		}
		return r < 32 || r >= 127
	}
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	str, _, _ = transform.String(t, str)
	return str
}

func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

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

// PrepareImage simply setups the image name
func PrepareImage(name, version string) string {
	return name + ":" + version
}

// PrepareCommand prepares the benchmark command
func PrepareCommand(command string) []string {
	return strings.Split(command, " ")
}

// PrepareBeforeCommands sets up before commands
// example output
// bash -c "command1 && command2"
func PrepareBeforeCommands(commands []string) []string {

	if len(commands) == 0 {
		return []string{}
	}

	prepared := strings.Join(commands, " && ")
	return []string{"bash", "-c", prepared}
}

func Welcome() {
	fmt.Printf("\n\r  %s\n\n", "ben started !")
}
