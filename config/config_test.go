package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ReadConfig(t *testing.T) {
	t.Run("missing ben.json", func(t *testing.T) {
		c, err := ReadConfig()
		assert.Nil(t, c)
		assert.NotNil(t, err)
	})
}

func TestConfig_Machines(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c := Config{
			Machines: []string{"s1", "s2"},
			Runtimes: []string{},
		}
		err := c.Validate()
		assert.Nil(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		c := Config{
			Machines: []string{"s1", "s9"},
			Runtimes: []string{},
		}
		err := c.Validate()
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "invalid machine size: s9")
	})
}
