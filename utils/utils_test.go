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
