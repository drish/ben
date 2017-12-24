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

func TestBuilder_LocalBuilder_SetupImage(t *testing.T) {

	t.Run("image does not exist", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:3.123123",
		}
		builder.Init()
		err := builder.SetupImage()
		assert.NotNil(t, err)
		assert.Equal(t, strings.Contains(err.Error(), "golang:3.123123 not found"), true)
	})

	t.Run("valid", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		builder.Init()
		err := builder.SetupImage()
		assert.Nil(t, err)
	})
}

func TestBuilder_LocalBuilder_SetupContainer(t *testing.T) {

	t.Run("image does not exist", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:2.999",
		}
		builder.Init()
		err := builder.SetupContainer()
		assert.NotNil(t, err)
		assert.Equal(t, strings.Contains(err.Error(), "Error: No such image: golang:2.999"), true)
	})

	t.Run("image exist", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		builder.Init()
		builder.SetupImage()
		err := builder.SetupContainer()
		assert.Nil(t, err)
	})
}

func TestBuilder_LocalBuilder_Cleanup(t *testing.T) {

	t.Run("sucessfull cleanup", func(t *testing.T) {

		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		builder.Init()
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
		assert.Equal(t, strings.Contains(err.Error(), "Container doesn't exist."), true)
	})
}
