package builders

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/drish/ben/utils"
	"github.com/fatih/color"
	hyper "github.com/hyperhq/hyper-api/client"
	"github.com/pkg/errors"
	spinner "github.com/tj/go-spin"
)

var (
	hosts = map[string]string{
		"us-west-1": "tcp://us-west-1.hyper.sh:443",
	}
	verStr = "v1.23"
)

// HyperBuilder is the Hyper.sh struct for dealing with hyper runtimes
type HyperBuilder struct {
	Image          string
	ID             string
	Name           string
	Size           string
	Before         []string
	Command        []string
	HyperClient    *hyper.Client
	DockerClient   *docker.Client
	BenchmarkImage string
}

// Init is a simple start message
func (b *HyperBuilder) Init() error {

	accessKey := os.Getenv("HYPER_ACCESSKEY")
	secretKey := os.Getenv("HYPER_SECRETKEY")
	region := os.Getenv("HYPER_REGION")

	if accessKey == "" || secretKey == "" {
		return errors.New("missing hyper.sh credentials")
	}

	// set default
	if region == "" {
		region = "us-west-1"
	}

	host := hosts[region]

	fmt.Printf("\r  \033[36msetting up environment on Hyper.sh %s for \033[m%s \n", region, b.Image)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	hyperClient, err := hyper.NewClient(host, verStr, httpClient, map[string]string{}, accessKey, secretKey, region)
	if err != nil {
		return errors.Wrap(err, "failed to connect setup hyper.sh client")
	}

	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "failed to connect to local docker")
	}

	b.DockerClient = dockerClient
	b.HyperClient = hyperClient
	return nil
}

// PrepareImage pulls the base image and run `before` commands
func (b *HyperBuilder) PrepareImage() error {

	if err := b.pullImage(); err != nil {
		return err
	}

	if err := b.setupBaseImage(); err != nil {
		return err
	}

	if err := b.runBeforeCommands(); err != nil {
		return err
	}

	if err := b.loadOnHyper(); err != nil {
		return err
	}

	if err := b.removeLocalImage(); err != nil {
		return err
	}

	return nil
}

// // SetupContainer creates the container on hyper
func (b *HyperBuilder) SetupContainer() error {
	return nil
}

// Benchmark runs the benchmark command
func (b *HyperBuilder) Benchmark() error {
	return nil
}

// Cleanup cleans up the environment on hyper
func (b *HyperBuilder) Cleanup() error {
	return nil
}

// Display writes the benchmark outputs to stdout
func (b *HyperBuilder) Display() error {
	return nil
}

// remove image from local fs
func (b *HyperBuilder) removeLocalImage() error {
	_, err := b.DockerClient.ImageRemove(context.Background(), b.BenchmarkImage, dockerTypes.ImageRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "failed removing benchmark image")
	}
	return nil
}

// load image on hyper.sh
func (b *HyperBuilder) loadOnHyper() error {

	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36muploading image to hyper.sh \033[m %s (%s)", color.MagentaString(s.Next()), "this may take a while.")
		}
		fmt.Printf("\r  \033[36muploading image to hyper.sh \033[m %s (%s)\n", color.GreenString("done !"), "this may take a while.")

	}()

	// craete image tar in order to be transferred to hyper
	cmd := []string{"docker", "save", "-o", b.BenchmarkImage + ".tar", b.BenchmarkImage}
	_, err := exec.Command(cmd[0], cmd[1], cmd[2], cmd[3], cmd[4]).Output()
	if err != nil {
		return errors.Wrap(err, "failed to create tar from image")
	}

	// NOTE: it is slow and not efficient to open a probably gb+ file like this
	// im not sure of an alternative atm
	tarFile, err := os.Open(b.BenchmarkImage + ".tar")
	if err != nil {
		return err
	}
	defer tarFile.Close()

	info, err := tarFile.Stat()
	if err != nil {
		return err
	}

	resp, err := b.HyperClient.ImageLoadLocal(context.Background(), true, info.Size())
	if err != nil {
		return err
	}
	defer resp.Conn.Close()

	_, err = io.Copy(resp.Conn, tarFile)
	if err != nil {
		return err
	}

	spin = false
	wg.Wait()
	return nil
}

// pull runtime image
func (b *HyperBuilder) pullImage() error {
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
	out, err := b.DockerClient.ImagePull(context.Background(), b.Image, dockerTypes.ImagePullOptions{})
	if err != nil {
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
func (b *HyperBuilder) setupBaseImage() error {

	config := &dockerContainer.Config{
		Image:      b.Image,
		WorkingDir: "/tmp",
		OpenStdin:  true,
	}

	// create tmp container
	tmpName := "ben-tmp-" + utils.RandString(8)
	c, err := b.DockerClient.ContainerCreate(context.Background(), config, nil, nil, tmpName)
	if err != nil {
		return errors.Wrap(err, "failed creating container")
	}

	// copy pwd data into tmp container
	cmd := []string{"docker", "cp", ".", c.ID + ":/tmp"}
	_, err = exec.Command(cmd[0], cmd[1], cmd[2], cmd[3]).Output()
	if err != nil {
		return errors.Wrap(err, "failed to copy data into container")
	}
	defer os.Remove(b.BenchmarkImage + ".tar")

	// create new image
	imageName := "ben-final-" + strings.ToLower(utils.RandString(4))
	_, err = b.DockerClient.ContainerCommit(context.Background(), c.ID, dockerTypes.ContainerCommitOptions{Reference: imageName})
	if err != nil {
		return errors.Wrap(err, "failed to create benchmark image")
	}

	// cleanup tmp container
	if err = b.removeContainer(c.ID); err != nil {
		return err
	}

	// save image name
	b.BenchmarkImage = imageName

	return nil
}

// run before commands if specified and create new image
func (b *HyperBuilder) runBeforeCommands() error {

	if len(b.Before) == 0 {
		fmt.Printf(" \033[36m no commands to run before !\n\033[m")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(1)

	tmpName := utils.RandString(10)

	config := &dockerContainer.Config{
		Image:      b.BenchmarkImage,
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        b.Before,
	}

	// create tmp container to run `before` commands
	c, err := b.DockerClient.ContainerCreate(context.Background(), config, nil, nil, tmpName)
	if err != nil {
		fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s ", color.RedString("failed !"))
		return errors.Wrap(err, "failed creating container")
	}

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)", color.MagentaString(s.Next()), strings.Join(b.Before, " "))
		}
	}()

	// start tmp container
	err = b.DockerClient.ContainerStart(context.Background(), c.ID, dockerTypes.ContainerStartOptions{})
	if err != nil {
		spin = false
		wg.Wait()
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exits
	exit, errC := b.DockerClient.ContainerWait(context.Background(), c.ID)
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

		fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)\n", color.RedString("failed !"), strings.Join(b.Before, " "))

		b.showOutput(c.ID)

		if err = b.removeContainer(c.ID); err != nil {
			return errors.Wrap(err, "failed removing tmp container")
		}

		_, err := b.DockerClient.ImageRemove(context.Background(), b.BenchmarkImage, dockerTypes.ImageRemoveOptions{Force: true, PruneChildren: true})
		if err != nil {
			return errors.Wrap(err, "failed removing benchmark image")
		}

		return errors.New("running 'before' commands failed")
	}

	oldImage := b.BenchmarkImage

	// create new image
	imageName := "ben-final-" + strings.ToLower(utils.RandString(4))
	_, err = b.DockerClient.ContainerCommit(context.Background(), c.ID, dockerTypes.ContainerCommitOptions{Reference: imageName})
	if err != nil {
		spin = false
		wg.Wait()
		return errors.Wrap(err, "failed to create benchmark image")
	}

	// save benchmark image name
	b.BenchmarkImage = imageName

	// cleanup tmp container
	if err = b.removeContainer(c.ID); err != nil {
		spin = false
		wg.Wait()

		return err
	}

	// cleanup previous image
	_, err = b.DockerClient.ImageRemove(context.Background(), oldImage, dockerTypes.ImageRemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		spin = false
		wg.Wait()

		return errors.Wrap(err, "failed removing benchmark image")
	}

	spin = false
	wg.Wait()

	fmt.Printf("\r  \033[36mrunning 'before' commands \033[m %s (%s)\n", color.GreenString("done !"), strings.Join(b.Before, " "))

	return nil
}

func (b *HyperBuilder) showOutput(containerID string) error {

	// // store container stdout
	// reader, err := l.Client.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{ShowStdout: true,
	// 	ShowStderr: true})
	// if err != nil {
	// 	return errors.Wrap(err, "failed to fetch logs")
	// }
	// defer reader.Close()

	// info, err := ioutil.ReadAll(reader)

	// fmt.Println()
	// fmt.Printf(string(info))

	return nil
}

func (b *HyperBuilder) removeContainer(containerID string) error {
	err := b.DockerClient.ContainerRemove(context.Background(), containerID, dockerTypes.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}
	return nil
}
