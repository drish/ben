package ben

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// RuntimeBuilder is the interface that defines how to build runtime environments
type RuntimeBuilder interface {
	PullImage() error
	SetupContainer() error
	Cleanup() error
}

// HyperBuilder is the Hyper.sh struct for dealing with hyper runtimes
type HyperBuilder struct {
	Image string
	ID    string
}

// LocalBuilder is the local struct for dealing with local runtimes
type LocalBuilder struct {
	Image string
	ID    string
}

// PullImage pulls the image on locally
func (l *LocalBuilder) PullImage() error {
	fmt.Println("Pulling the image locally:", l.Image)

	cli, err := client.NewEnvClient()

	if err != nil {
		return errors.New("failed to connect to local docker")
	}

	out, err := cli.ImagePull(context.Background(), l.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.New("failed pulling image")
	}

	rd := bufio.NewReader(out)

	for {
		// TODO: not read lines
		// not read lines, add spinner til done
		str, _, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.New("failed reading output")
		}
		fmt.Println(string(str))
	}
	return nil
}

// SetupContainer creates the container locally
func (l *LocalBuilder) SetupContainer() error {
	fmt.Println("Setting up container for:", l.Image)

	cli, err := client.NewEnvClient()

	if err != nil {
		return errors.New("failed to connect to local docker")
	}

	config := &container.Config{
		Image: l.Image,
		Volumes: map[string]struct{}{
			"/tmp": {},
		},
		OpenStdin: true,
	}

	bindPath, err := os.Getwd()
	if err != nil {
		return errors.New("failed to get current directory")
	}

	hostConfig := &container.HostConfig{
		Binds: []string{bindPath + ":/tmp"},
	}

	c, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, "namerino")

	if err != nil {
		return errors.New("failed creating container")
	}

	fmt.Println("Created container: ", c.ID)
	return nil
}

// Cleanup cleans up the environment
func (l *LocalBuilder) Cleanup() error {
	return nil
}

// PullImage pulls the image on hyper
func (b *HyperBuilder) PullImage() error {
	fmt.Println("Pulling the image on hyper.sh")
	return nil
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
