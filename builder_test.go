package ben

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_LocalBuilder_PullImage(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		builder := &LocalBuilder{
			Image: "golang:1.4",
		}
		err := builder.PullImage()
		assert.Nil(t, err)
	})
}

func TestBuilder_HyperBuilder_PullImage(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		builder := &HyperBuilder{
			Image: "golang:1.4",
		}
		err := builder.PullImage()
		assert.Nil(t, err)
	})
}
