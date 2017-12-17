package builders

import "fmt"

// HyperBuilder is the Hyper.sh struct for dealing with hyper runtimes
type HyperBuilder struct {
	Image string
	ID    string
	Name  string
	Size  string
}

// Init is a simple start message
func (b *HyperBuilder) Init() error {
	fmt.Printf("\r  \033[36mSetting up environment on Hyper.sh for \033[m%s \n", b.Image)
	return nil
}

// PullImage pulls the image on hyper
func (b *HyperBuilder) PullImage() error {
	return nil
}

// SetupContainer creates the container on hyper
func (b *HyperBuilder) SetupContainer() error {
	return nil
}

// Cleanup cleans up the environment on hyper
func (b *HyperBuilder) Cleanup() error {
	return nil
}
