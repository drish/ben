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
	command := "go test -v . ./config"

	c := PrepareCommand(command)
	assert.Equal(t, c[0], "go")
	assert.Equal(t, c[1], "test")
	assert.Equal(t, c[2], "-v")
	assert.Equal(t, c[3], ".")
	assert.Equal(t, c[4], "./config")
}
