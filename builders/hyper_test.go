package builders

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_HyperBuilder_Init(t *testing.T) {

	t.Run("env vars missing", func(t *testing.T) {

		os.Unsetenv("HYPER_ACCESSKEY")
		os.Unsetenv("HYPER_SECRETKEY")

		builder := &HyperBuilder{
			Image: "golang:1.4",
		}
		err := builder.Init()
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "missing hyper.sh credentials")
	})

	t.Run("env vars set", func(t *testing.T) {

		os.Setenv("HYPER_ACCESSKEY", "1")
		os.Setenv("HYPER_SECRETKEY", "2")

		builder := &HyperBuilder{
			Image: "golang:1.4",
		}
		err := builder.Init()
		assert.Nil(t, err)
		assert.NotNil(t, builder.HyperClient)
		assert.NotNil(t, builder.DockerClient)
	})
}
