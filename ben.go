package ben

import (
	"fmt"

	"github.com/drish/ben/config"
)

type Runner struct {
	config *config.Config
}

func (r *Runner) Run() error {

	fmt.Println("ben started !")

	for _, env := range r.config.Environments {
		image := env.Runtime + ":" + env.Version

		if env.Machine == "local" {
			builder := &LocalBuilder{
				Image: image,
			}
			r.BuildRuntime(builder)
		} else {
			builder := &HyperBuilder{
				Image: image,
			}
			r.BuildRuntime(builder)
		}
	}
	return nil
}

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

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}
