package ben

import (
	"fmt"

	"github.com/drish/ben/config"
)

// Runner defines the high-level runner struct
type Runner struct {
	config *config.Config
}

// Run is the entrypoint method
func (r *Runner) Run() error {

	fmt.Println("ben started !")

	for _, env := range r.config.Environments {
		image := env.Runtime + ":" + env.Version

		if env.Machine == "local" {
			builder := &LocalBuilder{
				Image: image,
			}
			if err := r.BuildRuntime(builder); err != nil {
				return err
			}
		} else {
			builder := &HyperBuilder{
				Image: image,
			}
			if err := r.BuildRuntime(builder); err != nil {
				return err
			}
		}
	}
	return nil
}

// BuildRuntime builds the appropriate runtime
func (r *Runner) BuildRuntime(b RuntimeBuilder) error {
	err := b.PullImage()
	if err != nil {
		return err
	}

	err = b.SetupContainer()
	if err != nil {
		return err
	}

	return nil
}

// New is the Runner initializer
func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}
