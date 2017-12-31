package builders

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/drish/ben/reporter"
	"github.com/drish/ben/utils"
	"github.com/fatih/color"
	hyper "github.com/hyperhq/hyper-api/client"
	hyperTypes "github.com/hyperhq/hyper-api/types"
	hyperContainer "github.com/hyperhq/hyper-api/types/container"
	"github.com/pkg/errors"
	spinner "github.com/tj/go-spin"
)

var (
	hosts = map[string]string{
		"us-west-1":    "tcp://us-west-1.hyper.sh:443",
		"eu-central-1": "tcp://eu-central-1.hyper.sh:443",
	}
	verStr = "v1.23"
)

var (
	sizesDescription = map[string]string{
		"s1": "s1 - 1 CPU 64MB",
		"s2": "s2 - 1 CPU 128MB",
		"s3": "s3 - 1 CPU 256MB",
		"s4": "s4 - 1 CPU 512MB",
		"m1": "m1 - 1 CPU 1GB",
		"m2": "m2 - 2 CPU 2GB",
		"m3": "m3 - 2 CPU 4GB",
		"l1": "l1 - 4 CPU 4GB",
		"l2": "l2 - 4 CPU 8GB",
		"l3": "l3 - 8 CPU 16GB",
	}
)

// HyperBuilder is the Hyper.sh struct for dealing with hyper runtimes
type HyperBuilder struct {
	Image          string // base image
	ID             string // benchmark container ID
	HyperSize      string
	Before         []string
	Command        []string
	Context        context.Context
	HyperClient    *hyper.Client
	HyperRegion    string
	DockerClient   *docker.Client
	BenchmarkImage string
	Results        string
}

// Init does requirements checks and sets up necessary variables
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

	if host == "" {
		return errors.New("invalid region set")
	}

	fmt.Printf("\r  \033[36msetting up environment on Hyper.sh %s for \033[m%s \n", region, b.Image)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	hyperClient, err := hyper.NewClient(host, verStr, httpClient, map[string]string{}, accessKey, secretKey, region)
	if err != nil {
		return errors.Wrap(err, "failed to setup hyper.sh client")
	}

	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "failed to connect to local docker")
	}

	b.DockerClient = dockerClient
	b.HyperClient = hyperClient
	b.HyperRegion = region
	b.Context = context.Background()
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

	if err := b.waitForImage(); err != nil {
		return err
	}

	if err := b.removeLocalImage(); err != nil {
		return err
	}

	return nil
}

// SetupContainer creates the container on hyper
func (b *HyperBuilder) SetupContainer() error {

	if b.Command == nil {
		b.Cleanup()
		return errors.New("command can not be blank")
	}

	if b.BenchmarkImage == "" {
		b.Cleanup()
		return errors.New("benchmark image not prepared")
	}

	config := &hyperContainer.Config{
		Image:      b.BenchmarkImage,
		WorkingDir: "/tmp",
		OpenStdin:  true,
		Cmd:        b.Command,
		Labels: map[string]string{
			"sh_hyper_instancetype": b.HyperSize,
		},
	}

	c, err := b.HyperClient.ContainerCreate(b.Context, config, nil, nil, "")
	if err != nil {
		b.Cleanup()
		fmt.Printf("\r  \033[36mcreating benchmark container \033[m %s ", color.RedString("failed !"))
		return errors.Wrap(err, "failed creating benchmark container")
	}

	fmt.Printf("  \033[36mcreating benchmark container \033[m %s (%s) \n", color.GreenString("done !"), sizesDescription[b.HyperSize])

	b.ID = c.ID
	return nil
}

// Benchmark runs the benchmark command on hyper
func (b *HyperBuilder) Benchmark() error {

	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("\r  \033[36mrunning benchmark \033[m %s (%s)", color.MagentaString(s.Next()), strings.Join(b.Command, " "))
		}
		fmt.Printf("\r  \033[36mrunning benchmark \033[m %s (%s)", color.GreenString("done !"), strings.Join(b.Command, " "))

	}()

	// start container
	err := b.HyperClient.ContainerStart(b.Context, b.ID, "")
	if err != nil {
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exits
	_, errC := b.HyperClient.ContainerWait(b.Context, b.ID)
	if err := errC; err != nil {
		return errors.Wrap(err, "failed to wait for container status")
	}

	// store container logs
	reader, err := b.HyperClient.ContainerLogs(b.Context, b.ID, hyperTypes.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return errors.Wrap(err, "failed to fetch logs")
	}

	results, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "failed to fetch logs")
	}

	b.Results = utils.StripCtlAndExtFromUnicode(string(results))
	spin = false

	wg.Wait()
	return nil
}

// Cleanup cleans up containers on hyper
func (b *HyperBuilder) Cleanup() error {

	fmt.Println()
	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mcleaning up container and volumes\033[m %s", color.MagentaString(s.Next()))
		}
		fmt.Printf("\r  \033[36mcleaning up container and volumes \033[m %s\n", color.GreenString("done !"))

	}()

	if b.ID == "" {
		return errors.New("container doesn't exist")
	}

	// try to remove benchmark container
	_, err := b.HyperClient.ContainerRemove(b.Context, b.ID, hyperTypes.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}

	// delete the image
	_, err = b.HyperClient.ImageRemove(b.Context, b.BenchmarkImage, hyperTypes.ImageRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "failed removing benchmark image")
	}

	spin = false
	wg.Wait()
	return nil
}

// Display writes the benchmark output to stdout
func (b *HyperBuilder) Display() error {
	fmt.Printf("  \033[36mdisplaying results\033[m \n")
	fmt.Println(b.Results)
	return nil
}

// Report returns data for being later written to fs
func (b *HyperBuilder) Report() reporter.ReportData {
	return reporter.ReportData{
		Image:   b.Image,
		Results: b.Results,
		Machine: "Hyper.sh cloud: " + sizesDescription[b.HyperSize],
		Before:  strings.Join(b.Before, " "),
		Command: strings.Join(b.Command, " "),
	}
}

// NOTE: ugly workaround, hyper takes a while to make the newly craeted image available.
// should be replaced by a checker to see if the image was uploaded every X secs
func (h *HyperBuilder) waitForImage() error {
	var wg sync.WaitGroup
	wg.Add(1)

	s := spinner.New()
	spin := true
	go func() {
		defer wg.Done()
		for spin == true {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("\r  \033[36mwaiting for image to become available \033[m %s", color.MagentaString(s.Next()))
		}
		fmt.Printf("\r  \033[36mwaiting for image to become available \033[m %s\n", color.GreenString("done !"))
	}()

	time.Sleep(20 * time.Second)
	spin = false

	wg.Wait()
	return nil
}

// remove image from local docker and fs
func (b *HyperBuilder) removeLocalImage() error {
	_, err := b.DockerClient.ImageRemove(b.Context, b.BenchmarkImage, dockerTypes.ImageRemoveOptions{})
	if err != nil {
		return errors.Wrap(err, "failed removing benchmark image")
	}

	err = os.Remove(b.BenchmarkImage + ".tar")
	if err != nil {
		return errors.Wrap(err, "unable to remove local image tar")
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

	resp, err := b.HyperClient.ImageLoadLocal(b.Context, true, info.Size())
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

// pull runtime base image
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

	// TODO: pull images from private repos
	out, err := b.DockerClient.ImagePull(b.Context, b.Image, dockerTypes.ImagePullOptions{})
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
	c, err := b.DockerClient.ContainerCreate(b.Context, config, nil, nil, tmpName)
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
	_, err = b.DockerClient.ContainerCommit(b.Context, c.ID, dockerTypes.ContainerCommitOptions{Reference: imageName})
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
	c, err := b.DockerClient.ContainerCreate(b.Context, config, nil, nil, tmpName)
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
	err = b.DockerClient.ContainerStart(b.Context, c.ID, dockerTypes.ContainerStartOptions{})
	if err != nil {
		spin = false
		wg.Wait()
		return errors.Wrap(err, "couldn't start container")
	}

	// wait until container exits
	exit, errC := b.DockerClient.ContainerWait(b.Context, c.ID)
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

		_, err := b.DockerClient.ImageRemove(b.Context, b.BenchmarkImage, dockerTypes.ImageRemoveOptions{Force: true, PruneChildren: true})
		if err != nil {
			return errors.Wrap(err, "failed removing benchmark image")
		}

		return errors.New("running 'before' commands failed")
	}

	oldImage := b.BenchmarkImage

	// create new image
	imageName := "ben-final-" + strings.ToLower(utils.RandString(4))
	_, err = b.DockerClient.ContainerCommit(b.Context, c.ID, dockerTypes.ContainerCommitOptions{Reference: imageName})
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
	_, err = b.DockerClient.ImageRemove(b.Context, oldImage, dockerTypes.ImageRemoveOptions{Force: true, PruneChildren: true})
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
	return nil
}

// Removes local container
func (b *HyperBuilder) removeContainer(containerID string) error {
	err := b.DockerClient.ContainerRemove(b.Context, containerID, dockerTypes.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return errors.Wrap(err, "failed removing container")
	}
	return nil
}
