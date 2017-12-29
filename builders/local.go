package builders

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/drish/ben/reporter"
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

// PrepareImage pulls the base image and run `before` commands
func (l *LocalBuilder) PrepareImage() error {

	if err := l.pullImage(); err != nil {
		return err
	}

	if err := l.setupBaseImage(); err != nil {
		return err
	}

	if err := l.runBeforeCommands(); err != nil {
		return err
	}

	return nil
}

// SetupContainer creates the final benchmark container locally
func (l *LocalBuilder) SetupContainer() error {

	if l.Command == nil {
		return errors.New("command can not be blank")
	}

	if l.BenchmarkImage == "" {
		return errors.New("benchmark image not prepared")
	}

	config := &container.Config{
		Image:      l.BenchmarkImage,
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        l.Command,
	}

	c, err := l.Client.ContainerCreate(context.Background(), config, nil, nil, "")
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

	// store container logs
	// using exec here because was having problems with encoding on ContainerLogs
	cmd := []string{"docker", "logs", l.ID}
	out, err := exec.Command(cmd[0], cmd[1], cmd[2]).Output()
	if err != nil {
		return errors.Wrap(err, "failed to copy data into container")
	}

	l.Results = string(out)
	spin = false
	s.Reset()

	wg.Wait()
	return nil
}

// Cleanup cleans up containers used for benchmarking
func (l *LocalBuilder) Cleanup() error {

	if l.ID == "" {
		return errors.New("container doesn't exist")
	}

	// try to remove benchmark container
	err := l.Client.ContainerRemove(context.Background(), l.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}

	// delete the image
	_, err = l.Client.ImageRemove(context.Background(), l.BenchmarkImage, types.ImageRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "failed removing benchmark image")
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

// Report returns data for being later written to fs
func (l *LocalBuilder) Report() reporter.ReportData {
	d := reporter.ReportData{
		Image:   l.Image,
		Results: l.Results,
		Machine: "local",
		Command: strings.Join(l.Command, " "),
	}
	return d
}

// pull runtime image
func (l *LocalBuilder) pullImage() error {
	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mpreparing image \033[m %s", color.MagentaString(s.Next()))
		}
		fmt.Printf("\r  \033[36mpreparing image \033[m %s\n", color.GreenString("done !"))

	}()

	// pulls runtime image
	// TODO: pull images from private repos
	out, err := l.Client.ImagePull(context.Background(), l.Image, types.ImagePullOptions{})
	if err != nil {
		fmt.Printf("\r  \033[36mpreparing image \033[m %s\n", color.RedString("failed !"))
		return errors.Wrap(err, "failed preparing image")
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
			fmt.Printf("\r  \033[36mpreparing image \033[m %s\n", color.RedString("failed !"))
			return errors.New("failed reading output")
		}
	}
}

// setup working dir and copy pwd dir into
func (l *LocalBuilder) setupBaseImage() error {

	config := &container.Config{
		Image:      l.Image,
		WorkingDir: "/tmp",
		OpenStdin:  true,
	}

	// create tmp container
	tmpName := "ben-tmp-" + utils.RandString(8)
	c, err := l.Client.ContainerCreate(context.Background(), config, nil, nil, tmpName)
	if err != nil {
		return errors.Wrap(err, "failed creating container")
	}

	// copy pwd data into tmp container
	cmd := []string{"docker", "cp", ".", c.ID + ":/tmp"}
	_, err = exec.Command(cmd[0], cmd[1], cmd[2], cmd[3]).Output()
	if err != nil {
		return errors.Wrap(err, "failed to copy data into container")
	}

	// create new image
	imageName := "ben-final-" + strings.ToLower(utils.RandString(4))
	_, err = l.Client.ContainerCommit(context.Background(), c.ID, types.ContainerCommitOptions{Reference: imageName})
	if err != nil {
		return errors.Wrap(err, "failed to create benchmark image")
	}

	// cleanup tmp container
	if err = l.removeContainer(c.ID); err != nil {
		return err
	}

	// save image name
	l.BenchmarkImage = imageName

	return nil
}

// run before commands if specified and create new image
func (l *LocalBuilder) runBeforeCommands() error {

	if len(l.Before) == 0 {
		fmt.Printf(" \033[36m no commands to run before !\n\033[m")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(1)

	tmpName := utils.RandString(10)

	config := &container.Config{
		Image:      l.BenchmarkImage,
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        l.Before,
	}

	// create tmp container to run `before` commands
	c, err := l.Client.ContainerCreate(context.Background(), config, nil, nil, tmpName)
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
	}()

	// start tmp container
	err = l.Client.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
	if err != nil {
		spin = false
		wg.Wait()
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exits
	exit, errC := l.Client.ContainerWait(context.Background(), c.ID)
	if err := errC; err != nil {
		spin = false
		wg.Wait()
		return errors.Wrap(err, "failed to wait for container status")
	}

	// if some command fails on tmp container
	// shows error outputs
	if exit != 0 {

		spin = false
		wg.Wait()

		fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)\n", color.RedString("failed !"), strings.Join(l.Before, " "))

		l.showOutput(c.ID)

		if err = l.removeContainer(c.ID); err != nil {
			return errors.Wrap(err, "failed removing tmp container")
		}

		_, err := l.Client.ImageRemove(context.Background(), l.BenchmarkImage, types.ImageRemoveOptions{})
		if err != nil {
			return errors.Wrap(err, "failed removing benchmark image")
		}

		return errors.New("running 'before' commands failed")
	}

	oldImage := l.BenchmarkImage

	// create new image
	imageName := "ben-final-" + strings.ToLower(utils.RandString(4))
	_, err = l.Client.ContainerCommit(context.Background(), c.ID, types.ContainerCommitOptions{Reference: imageName})
	if err != nil {
		spin = false
		wg.Wait()
		return errors.Wrap(err, "failed to create benchmark image")
	}

	// save benchmark image name
	l.BenchmarkImage = imageName

	// cleanup tmp container
	if err = l.removeContainer(c.ID); err != nil {
		spin = false
		wg.Wait()

		return err
	}

	// cleanup previous image
	_, err = l.Client.ImageRemove(context.Background(), oldImage, types.ImageRemoveOptions{})
	if err != nil {
		spin = false
		wg.Wait()

		return errors.Wrap(err, "failed removing benchmark image")
	}

	spin = false
	wg.Wait()

	fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)\n", color.GreenString("done !"), strings.Join(l.Before, " "))

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
