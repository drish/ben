package ben

import (
	"fmt"

	"github.com/drish/ben/config"
)

type Runner struct {
	config *config.Config
}

func (r *Runner) Run() error {
	defer fmt.Println()
	fmt.Println("ben started !")
	err := r.BuildRuntimes()
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) BuildRuntimes() error {
	fmt.Println(r.config)
	// for _, r := range r.config. {
	// 	fmt.Println("builing : ", r)
	// }
	return nil
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}
