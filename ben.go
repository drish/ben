package ben

import (
	"fmt"

	"github.com/drish/ben/builders"
	"github.com/drish/ben/config"
)

// Runner defines the high-level runner struct
type Runner struct {
	config *config.Config
}

// Run is the entrypoint method
func (r *Runner) Run() error {

	fmt.Printf("\n\r  ben started ! \n\n")

	for _, env := range r.config.Environments {
		image := env.Runtime + ":" + env.Version

		if env.Machine == "local" {
			builder := &builders.LocalBuilder{
				Image: image,
			}
			if err := r.BuildRuntime(builder); err != nil {
				return err
			}
		} else {
			builder := &builders.HyperBuilder{
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
func (r *Runner) BuildRuntime(b builders.RuntimeBuilder) error {

	displayResults := true

	err := b.Init()
	if err != nil {
		return err
	}

	err = b.PullImage()
	if err != nil {
		return err
	}

	err = b.SetupContainer()
	if err != nil {
		return err
	}

	err = b.Benchmark()
	if err != nil {
		return err
	}

	err = b.Cleanup()
	if err != nil {
		return err
	}

	if displayResults {
		b.Display()
	}

	return nil
}

// New is the Runner initializer
func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}
