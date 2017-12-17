package builders

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	spin "github.com/tj/go-spin"
)

// LocalBuilder is the local struct for dealing with local runtimes
type LocalBuilder struct {
	Image   string
	ID      string
	Name    string
	Client  *client.Client
	Results string
}

// Init initializes necessary vars
func (l *LocalBuilder) Init() error {

	fmt.Printf("\r  \033[36msetting up local environment for \033[m%s \n", l.Image)

	cli, err := client.NewEnvClient()

	if err != nil {
		return errors.Wrap(err, "failed to connect to local docker")
	}

	l.Client = cli
	return nil
}

// PullImage pulls the image on locally
func (l *LocalBuilder) PullImage() error {

	// TODO: pull images from private repos
	out, err := l.Client.ImagePull(context.Background(), l.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.New("failed pulling image")
	}

	rd := bufio.NewReader(out)

	s := spin.New()
	for {
		_, _, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				s.Reset()
				fmt.Printf("\r  \033[36mpreparing image \033[m %s \n", color.GreenString("done !"))
				return nil
			}
			s.Reset()
			fmt.Printf("\r  \033[36mpreparing image \033[m %s ", color.RedString("failed !"))
			return errors.New("failed reading output")
		}
		fmt.Printf("\r  \033[36mpreparing image \033[m %s ", s.Next())
	}
}

// SetupContainer creates the container locally
func (l *LocalBuilder) SetupContainer() error {

	config := &container.Config{
		Image: l.Image,
		Volumes: map[string]struct{}{
			"/tmp": {},
		},
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        []string{"go", "test", "-bench=."},
	}

	bindPath, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current directory")
	}

	// binds pwd into the container /tmp
	hostConfig := &container.HostConfig{
		Binds: []string{bindPath + ":/tmp"},
	}

	c, err := l.Client.ContainerCreate(context.Background(), config, hostConfig, nil, "")
	if err != nil {
		fmt.Printf("\r  \033[36mcreating container \033[m %s ", color.RedString("failed !"))
		return errors.Wrap(err, "failed creating container")
	}

	fmt.Printf("\r  \033[36mcreated container \033[m %s %s \n", c.ID[:10], color.GreenString("done !"))
	l.ID = c.ID
	return nil
}

func (l *LocalBuilder) Benchmark() error {

	defer fmt.Println()

	s := spin.New()
	spin := true

	go func() {
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mrunning benchmark \033[m %s", s.Next())
		}
	}()

	// start container
	err := l.Client.ContainerStart(context.Background(), l.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exists
	_, errC := l.Client.ContainerWait(context.Background(), l.ID)
	if err := errC; err != nil {
		log.Fatal(err)
	}

	spin = false
	fmt.Printf("\r  \033[36mrunning benchmark \033[m %s", color.GreenString("done !"))

	// store container logs
	reader, err := l.Client.ContainerLogs(context.Background(), l.ID, types.ContainerLogsOptions{ShowStdout: true,
		ShowStderr: true})

	if err != nil {
		return errors.Wrap(err, "failed to fetch logs")
	}

	info, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "failed to fetch logs")
	}
	l.Results = string(info)
	return nil
}

// Cleanup cleans up containers used for benchmarking
func (l *LocalBuilder) Cleanup() error {

	err := l.Client.ContainerRemove(context.Background(), l.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}

	fmt.Printf("\r  \033[36mremoving container \033[m %s %s \n", l.ID[:10], color.GreenString("done !"))
	return nil
}

func (l *LocalBuilder) Display() error {
	fmt.Printf("\r  \033[36mdisplaying results\033[m \n")
	fmt.Println(l.Results)
	return nil
}
