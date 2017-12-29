package ben

import (
	"errors"
	"fmt"

	"github.com/drish/ben/builders"
	"github.com/drish/ben/config"
	"github.com/drish/ben/reporter"
	"github.com/drish/ben/utils"
)

// Runner defines the top-level runner struct
type Runner struct {
	config *config.Config
}

// Run is the entrypoint method
func (r *Runner) Run(output string, display bool) error {

	utils.Welcome()

	var reports []reporter.ReportData

	for _, env := range r.config.Environments {

		// set version as latest if no set
		if env.Version == "" {
			env.Version = "latest"
		}

		// set default command or exit
		if env.Command == "" {
			defaultCmd := config.DefaultCommand(env.Runtime)
			if defaultCmd == "" {
				return errors.New("command can not be blank")
			}
			env.Command = defaultCmd
		}

		before := utils.PrepareBeforeCommands(env.Before)
		image := utils.PrepareImage(env.Runtime, env.Version)

		command := utils.PrepareCommand(env.Command)

		if env.Machine == "local" {
			builder := &builders.LocalBuilder{
				Image:   image,
				Before:  before,
				Command: command,
			}
			rp, err := r.BuildRuntime(builder, output, display)
			if err != nil {
				return err
			}
			reports = append(reports, rp)
		} else {
			builder := &builders.HyperBuilder{
				Image:   image,
				Before:  before,
				Command: command,
			}
			rp, err := r.BuildRuntime(builder, output, display)
			if err != nil {
				return err
			}
			reports = append(reports, rp)
		}
	}

	// generate reports
	rep := reporter.NewReporter(output)
	if err := rep.Run(reports); err != nil {
		return err
	}
	return nil
}

// BuildRuntime builds the appropriate runtime
func (r *Runner) BuildRuntime(b builders.RuntimeBuilder, output string, display bool) (reporter.ReportData, error) {

	// sets up necessary variables
	if err := b.Init(); err != nil {
		return reporter.ReportData{}, err
	}

	// pulls base image, run before commands and create benchmark image
	if err := b.PrepareImage(); err != nil {
		return reporter.ReportData{}, err
	}

	if err := b.SetupContainer(); err != nil {
		return reporter.ReportData{}, err
	}

	if err := b.Benchmark(); err != nil {
		return reporter.ReportData{}, err
	}

	if err := b.Cleanup(); err != nil {
		return reporter.ReportData{}, err
	}

	if display {
		b.Display()
	} else {
		fmt.Println()
	}

	return b.Report(), nil
}

// New is the Runner initializer
func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}
