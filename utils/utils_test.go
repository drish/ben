package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	list := []string{
		"123",
		"aa",
	}
	assert.Equal(t, Contains("aa", list), true)
}

func TestPrepareCommand(t *testing.T) {

	t.Run("command is empty", func(t *testing.T) {
		command := ""

		c := PrepareCommand(command)
		assert.Equal(t, len(c), 1)
		assert.Equal(t, c, []string{""})
	})

	t.Run("command is not empty", func(t *testing.T) {
		command := "go test -v -bench=."

		c := PrepareCommand(command)
		assert.Equal(t, len(c), 4)
		assert.Equal(t, c, []string{"go", "test", "-v", "-bench=."})
	})
}

func TestPrepareImage(t *testing.T) {
	assert.Equal(t, PrepareImage("golang", "1.4"), "golang:1.4")
}

func TestPrepareBeforeCommands(t *testing.T) {

	t.Run("simple single command", func(t *testing.T) {
		c := PrepareBeforeCommands([]string{"apt-get update"})
		assert.Equal(t, c, []string{"bash", "-c", "apt-get update"})
	})

	t.Run("multiple commands", func(t *testing.T) {

		c := PrepareBeforeCommands([]string{"apt-get update", "echo test", "ls"})
		assert.Equal(t, c, []string{"bash", "-c", "apt-get update && echo test && ls"})
	})
}
