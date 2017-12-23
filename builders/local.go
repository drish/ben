package builders

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	spinner "github.com/tj/go-spin"
)

// LocalBuilder is the local struct for dealing with local runtimes
type LocalBuilder struct {
	Image   string
	Command []string
	ID      string
	Name    string
	Client  *client.Client
	Results string
}

// Init initializes necessary variables
func (l *LocalBuilder) Init() error {

	fmt.Printf("  \033[36msetting up local environment for \033[m%s \n", l.Image)

	cli, err := client.NewEnvClient()

	if err != nil {
		return errors.Wrap(err, "failed to connect to local docker")
	}

	l.Client = cli
	return nil
}

// PullImage pulls the image on locally
func (l *LocalBuilder) PullImage() error {

	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mpulling image \033[m %s", color.MagentaString(s.Next()))
		}
		fmt.Printf("\r  \033[36mpulling image \033[m %s\n", color.GreenString("done !"))

	}()

	// TODO: pull images from private repos
	out, err := l.Client.ImagePull(context.Background(), l.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "failed pulling image")
	}

	rd := bufio.NewReader(out)
	for {
		_, _, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				s.Reset()
				spin = false
				wg.Wait()
				return nil
			}
			s.Reset()
			spin = false
			fmt.Printf("\r  \033[36mpulling image \033[m %s\n", color.RedString("failed !"))
			return errors.New("failed reading output")
		}
	}
}

// SetupContainer creates the container locally
// it setups a volume with a local dir to the /tmp in the container
func (l *LocalBuilder) SetupContainer() error {

	config := &container.Config{
		Image: l.Image,
		Volumes: map[string]struct{}{
			"/tmp": {},
		},
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        l.Command,
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

	fmt.Printf("  \033[36mcreating container \033[m %s (%s) \n", color.GreenString("done !"), c.ID[:10])
	l.ID = c.ID
	return nil
}

// Benchmark runs the benchmark command
func (l *LocalBuilder) Benchmark() error {

	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("\r  \033[36mrunning benchmark \033[m %s (%s)", color.MagentaString(s.Next()), strings.Join(l.Command, " "))
		}
		fmt.Printf("\r  \033[36mrunning benchmark \033[m %s (%s)", color.GreenString("done !"), strings.Join(l.Command, " "))

	}()

	// start container
	err := l.Client.ContainerStart(context.Background(), l.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exists
	_, errC := l.Client.ContainerWait(context.Background(), l.ID)
	if err := errC; err != nil {
		return errors.Wrap(err, "failed to wait for container status")
	}

	// store container stdout
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
	spin = false
	s.Reset()

	wg.Wait()
	return nil
}

// Cleanup cleans up containers used for benchmarking
func (l *LocalBuilder) Cleanup() error {

	if l.ID == "" {
		return errors.New("Container doesn't exist.")
	}
	// try to remove container
	err := l.Client.ContainerRemove(context.Background(), l.ID, types.ContainerRemoveOptions{RemoveVolumes: true})

	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}

	fmt.Println()
	fmt.Printf("  \033[36mcleaning up container and volumes\033[m %s \n", color.GreenString(" done !"))
	return nil
}

// Display writes the benchmark output to stdout
func (l *LocalBuilder) Display() error {
	fmt.Printf("  \033[36mdisplaying results\033[m \n")
	fmt.Println(l.Results)
	return nil
}
