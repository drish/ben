package builders

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_LocalBuilder_Init(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		err := builder.Init()
		assert.Nil(t, err)
		assert.NotNil(t, builder.Client)
	})
}

func TestBuilder_LocalBuilder_PrepareImage(t *testing.T) {

	t.Run("image does not exist", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:3.123123",
		}
		builder.Init()
		err := builder.PrepareImage()
		assert.NotNil(t, err)
		assert.Equal(t, strings.Contains(err.Error(), "golang:3.123123 not found"), true)
		builder.Cleanup()
	})

	t.Run("valid", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		builder.Init()
		err := builder.PrepareImage()
		assert.Nil(t, err)
		builder.Cleanup()
	})
}

func TestBuilder_LocalBuilder_SetupContainer(t *testing.T) {

	t.Run("image does not exist", func(t *testing.T) {
		builder := &LocalBuilder{
			Image:   "golang:2.999",
			Command: []string{"ls", "-lah"},
		}
		builder.Init()
		err := builder.SetupContainer()
		assert.NotNil(t, err)
		assert.Equal(t, strings.Contains(err.Error(), "benchmark image not prepared"), true)
		builder.Cleanup()
	})

	t.Run("image exist", func(t *testing.T) {
		builder := &LocalBuilder{
			Image:   "golang:1.4",
			Command: []string{"ls", "-lah"},
		}
		builder.Init()
		builder.PrepareImage()
		err := builder.SetupContainer()
		assert.Nil(t, err)
	})
}

func TestBuilder_LocalBuilder_Cleanup(t *testing.T) {

	t.Run("sucessfull cleanup", func(t *testing.T) {

		builder := &LocalBuilder{
			Image:   "golang:1.4",
			Command: []string{"ls", "-lah"},
		}
		builder.Init()
		builder.PrepareImage()
		builder.SetupContainer()
		err := builder.Cleanup()
		assert.Nil(t, err)
	})

	t.Run("failed cleanup", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}

		err := builder.Cleanup()
		assert.NotNil(t, err)
		assert.Equal(t, strings.Contains(err.Error(), "container doesn't exist"), true)
	})
}
