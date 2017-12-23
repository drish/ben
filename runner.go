package ben

import (
	"fmt"

	"github.com/drish/ben/builders"
	"github.com/drish/ben/config"
	"github.com/drish/ben/utils"
)

// Runner defines the top-level runner struct
type Runner struct {
	config *config.Config
}

// Run is the entrypoint method
func (r *Runner) Run(output string, display bool) error {

	fmt.Printf("\n\r  %s\n\n", "ben started !")

	for _, env := range r.config.Environments {
		image := utils.PrepareImage(env.Runtime, env.Version)
		command := utils.PrepareCommand(env.Command)

		if env.Machine == "local" {
			builder := &builders.LocalBuilder{
				Image:   image,
				Command: command,
			}
			if err := r.BuildRuntime(builder, output, display); err != nil {
				return err
			}
		} else {
			builder := &builders.HyperBuilder{
				Image:   image,
				Command: command,
			}
			if err := r.BuildRuntime(builder, output, display); err != nil {
				return err
			}
		}
	}
	return nil
}

// BuildRuntime builds the appropriate runtime
func (r *Runner) BuildRuntime(b builders.RuntimeBuilder, output string, display bool) error {

	if err := b.Init(); err != nil {
		return err
	}

	if err := b.PullImage(); err != nil {
		return err
	}

	if err := b.SetupContainer(); err != nil {
		return err
	}

	if err := b.Benchmark(); err != nil {
		return err
	}

	if err := b.Cleanup(); err != nil {
		return err
	}

	if display {
		b.Display()
	} else {
		fmt.Println()
	}

	return nil
}

// New is the Runner initializer
func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}
