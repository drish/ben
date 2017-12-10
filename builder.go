package ben

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// interface that defines how to pull images
type RuntimeBuilder interface {
	PullImage() error
}

// builds the image on azure
type AzureBuilder struct {
}

type HyperBuilder struct {
	Image string
	ID    string
}

type LocalBuilder struct {
	Image string
	ID    string
}

// builds the image locally
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

// builds the image on hyper
func (b *HyperBuilder) PullImage() error {
	fmt.Println("Pulling the image on hyper.sh")
	return nil
}
