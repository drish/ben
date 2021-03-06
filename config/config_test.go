package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ReadConfig(t *testing.T) {

	t.Run("missing ben.json", func(t *testing.T) {
		c, err := ReadConfig("a")
		assert.Nil(t, c)
		assert.NotNil(t, err)
	})
}

func TestConfig_Runtime(t *testing.T) {

	t.Run("runtime is blank", func(t *testing.T) {
		e := Environment{
			Version: "1.9",
			Machine: "hyper-s4",
		}
		c := Config{
			Environments: []Environment{e},
		}
		err := c.Validate()
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "environment 0 runtime can't be blank")
	})
}

func TestConfig_Machines(t *testing.T) {

	t.Run("valid", func(t *testing.T) {

		e := Environment{
			Version: "1.9",
			Runtime: "golang",
			Machine: "hyper-s4",
		}
		c := Config{
			Environments: []Environment{e},
		}
		err := c.Validate()
		assert.Nil(t, err)
	})

	t.Run("invalid", func(t *testing.T) {

		e := Environment{
			Version: "1.9",
			Runtime: "golang",
			Machine: "s9",
		}

		c := Config{
			Environments: []Environment{e},
		}
		err := c.Validate()
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "invalid machine size: s9")
	})
}

func TestConfig_DefaultCommand(t *testing.T) {
	command := DefaultCommand("golang")
	assert.Equal(t, command, "go test -bench=.")
}
