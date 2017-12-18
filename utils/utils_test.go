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

	t.Run("command is empty should default to base command", func(t *testing.T) {
		command := ""

		defaultCommand := []string{"go", "test", "-bench=."}

		c := PrepareCommand(command)
		assert.Equal(t, c, defaultCommand)
	})

	t.Run("command is not empty", func(t *testing.T) {
		command := "go test -v . ./config"

		c := PrepareCommand(command)
		assert.Equal(t, c[0], "go")
		assert.Equal(t, c[1], "test")
		assert.Equal(t, c[2], "-v")
		assert.Equal(t, c[3], ".")
		assert.Equal(t, c[4], "./config")
	})
}

func TestPrepareImage(t *testing.T) {
	assert.Equal(t, PrepareImage("golang", "1.4"), "golang:1.4")
}
