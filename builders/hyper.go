package builders

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/hyperhq/hyper-api/client"
	"github.com/hyperhq/hyper-api/types"
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
	Image       string
	ID          string
	Name        string
	Size        string
	Before      []string
	Command     []string
	HyperClient *client.Client
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

	fmt.Printf("\r  \033[36mSetting up environment on Hyper.sh %s for \033[m%s \n", region, b.Image)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	client, err := client.NewClient(host, verStr, httpClient, map[string]string{}, accessKey, secretKey, region)
	if err != nil {
		return err
	}

	b.HyperClient = client
	return nil
}

// PullImage pulls the image on hyper
func (b *HyperBuilder) SetupImage() error {

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
	out, err := b.HyperClient.ImagePull(context.Background(), b.Image, types.ImagePullOptions{})
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

// RunBefore run before commands if specified
func (b *HyperBuilder) RunBefore() error {
	return nil
}

// SetupContainer creates the container on hyper
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
