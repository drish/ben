package builders

import (
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

func TestBuilder_LocalBuilder_PullImage(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		builder.Init()
		err := builder.PullImage()
		assert.Nil(t, err)
	})
}

func TestBuilder_LocalBuilder_SetupContainer(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		builder.Init()
		err := builder.SetupContainer()
		assert.Nil(t, err)
	})
}

func TestBuilder_LocalBuilder_Cleanup(t *testing.T) {

	builder := &LocalBuilder{
		Image: "golang:1.4",
	}
	builder.Init()
	builder.SetupContainer()

	err := builder.Cleanup()
	assert.Nil(t, err)
}
