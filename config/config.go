package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/drish/ben/utils"
	"github.com/pkg/errors"
)

// in case `command` is left blank on json file
// set a default command
var defaultCommands = map[string]string{
	"golang": "go test -bench=.",
}

// Machine sizes, both for local runtime and for Hyper.sh
var machineSizes = []string{
	"hyper-s1", // 1 CPU  64MB
	"hyper-s2", // 1 CPU  128MB
	"hyper-s3", // 1 CPU  256MB
	"hyper-s4", // 1 CPU  512MB
	"hyper-m1", // 1 CPU  1GB
	"hyper-m2", // 2 CPUs 2GB
	"hyper-m3", // 2 CPUs 4GB
	"hyper-l1",
	"hyper-l2",
	"hyper-l3",
	"local",
}

// representation of json config file
type Environment struct {
	Machine string   // hyper.sh machine size, ie: s1
	Version string   // runtime version, ie 1.9
	Runtime string   // runtime name, ie: golang, ruby, jruby
	Command string   // benchmark command
	Before  []string // commands to run on container before benchmark is done
}

type Config struct {
	Environments []Environment `json:"environments"`
}

// checks if provided machine size is on list of supported sizes
func validateMachineSizes(sizes []string) error {
	for _, s := range sizes {
		if !utils.Contains(s, machineSizes) {
			return errors.New("invalid machine size " + s)
		}
	}
	return nil
}

// validates all configuration provided
func (c *Config) Validate() error {

	// validates machine sizes
	var sizes []string
	for _, env := range c.Environments {
		sizes = append(sizes, env.Machine)
	}
	if err := validateMachineSizes(sizes); err != nil {
		return err
	}

	return nil
}

// parses json bytes into a Config struct
func ParseConfig(b []byte) (*Config, error) {
	c := &Config{}

	if err := json.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "unmarshalling error")
	}

	if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation error")
	}

	return c, nil
}

// reads the file at `path`
func ReadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "file open failed")
	}
	return ParseConfig(b)
}

// DefaultCommand returns the default command for the specified runtime
func DefaultCommand(runtime string) string {
	return defaultCommands[runtime]
}
