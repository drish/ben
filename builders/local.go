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
	"github.com/drish/ben/utils"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	spinner "github.com/tj/go-spin"
)

// LocalBuilder is the local struct for managing with local runtimes
type LocalBuilder struct {
	Image          string         // runtime base image
	Command        []string       // benchmark command
	Before         []string       // commands to run before bench
	ID             string         // benchmark container id
	Client         *client.Client // docker client
	Results        string         // benchmark output
	BenchmarkImage string         // if `before` is set a new image is created
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

// SetupImage pulls the runtime image on locally
// If `Before` is set, runs `Before` commands and create a new benchmark image
func (l *LocalBuilder) SetupImage() error {

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

// RunBefore runs commands specificed on `Before` field on the json config file
// creates a tmp container to run commands
// and creates a new image based on that
func (l *LocalBuilder) RunBefore() error {

	if len(l.Before) == 0 {
		fmt.Printf(" \033[36m no commands to run before !\n\033[m")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(1)

	tmpName := utils.RandString(10)

	config := &container.Config{
		Image: l.Image,
		Volumes: map[string]struct{}{
			"/tmp": {},
		},
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        l.Before,
	}

	bindPath, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current directory")
	}

	// binds pwd into the container /tmp
	hostConfig := &container.HostConfig{
		Binds: []string{bindPath + ":/tmp"},
	}

	// create tmp container to run `before` commands
	c, err := l.Client.ContainerCreate(context.Background(), config, hostConfig, nil, tmpName)
	if err != nil {
		fmt.Printf("\r  \033[36mrunning before commands \033[m %s ", color.RedString("failed !"))
		return errors.Wrap(err, "failed creating container")
	}

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)", color.MagentaString(s.Next()), strings.Join(l.Before, " "))
		}
		fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)\n", color.GreenString("done !"), strings.Join(l.Before, " "))
	}()

	// start tmp container
	err = l.Client.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exits
	exit, errC := l.Client.ContainerWait(context.Background(), c.ID)
	if err := errC; err != nil {
		return errors.Wrap(err, "failed to wait for container status")
	}

	// if some command fails on tmp container
	// shows error outputs
	if exit != 0 {
		fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)\n", color.RedString("failed !"), strings.Join(l.Before, " "))
		l.showOutput(c.ID)

		if err = l.removeContainer(c.ID); err != nil {
			return err
		}
		return errors.New("'before' commands failed !")
	}

	imageName := "ben-final-" + strings.ToLower(utils.RandString(4))
	_, err = l.Client.ContainerCommit(context.Background(), c.ID, types.ContainerCommitOptions{Reference: imageName})
	if err != nil {
		return errors.Wrap(err, "failed to create benchmark image")
	}

	// save benchmark image name
	l.BenchmarkImage = imageName

	// cleanup tmp container
	if err = l.removeContainer(c.ID); err != nil {
		return err
	}

	spin = false
	wg.Wait()
	return nil
}

// SetupContainer creates the final benchmark container locally
// it setups a volume with a local dir to the /tmp in the container
func (l *LocalBuilder) SetupContainer() error {

	var image string

	// if before was set
	// use the new prepared image as the benchmark container image
	if len(l.Before) == 0 {
		image = l.Image
	} else {
		image = l.BenchmarkImage
	}

	config := &container.Config{
		Image: image,
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
		fmt.Printf("\r  \033[36mcreating benchmark container \033[m %s ", color.RedString("failed !"))
		return errors.Wrap(err, "failed creating benchmark container")
	}

	fmt.Printf("  \033[36mcreating benchmark container \033[m %s (%s) \n", color.GreenString("done !"), c.ID[:10])
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

	// wait until container exits
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
	// try to remove benchmark container
	err := l.Client.ContainerRemove(context.Background(), l.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}

	// delete the bechmark image if it was created
	if l.BenchmarkImage != "" {
		_, err := l.Client.ImageRemove(context.Background(), l.BenchmarkImage, types.ImageRemoveOptions{})
		if err != nil {
			return errors.Wrap(err, "failed removing benchmark image")
		}
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

// writes container logs to stdout
// TODO: print better
func (l *LocalBuilder) showOutput(containerID string) error {

	// store container stdout
	reader, err := l.Client.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{ShowStdout: true,
		ShowStderr: true})
	if err != nil {
		return errors.Wrap(err, "failed to fetch logs")
	}
	defer reader.Close()

	info, err := ioutil.ReadAll(reader)

	fmt.Println()
	fmt.Printf(string(info))

	return nil
}

func (l *LocalBuilder) removeContainer(containerID string) error {
	err := l.Client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}
	return nil
}
