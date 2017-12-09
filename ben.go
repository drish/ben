package ben

import "fmt"

type Runner struct{}

func (r *Runner) Run() {
	r.run()
}

func (r *Runner) run() {
	fmt.Println("ben started")
}

func New() *Runner {
	return &Runner{}
}
