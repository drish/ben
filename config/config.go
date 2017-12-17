package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/drish/ben/utils"
	"github.com/pkg/errors"
)

// Runtimes ben is able to parse
var supportedRuntimes = []string{
	"golang",
}

// Machine sizes, both for local runtimes and for Hyper.sh
var machineSizes = []string{
	"s1", // 1 CPU  64MB
	"s2", // 1 CPU  128MB
	"s3", // 1 CPU  256MB
	"s4", // 1 CPU  512MB
	"m1", // 1 CPU  1GB
	"m2", // 2 CPUs 2GB
	"m3", // 2 CPUs 4GB
	"l1",
	"l2",
	"l3",
	"local",
}

// representation of json config file
type Environment struct {
	Machine string
	Version string
	Runtime string
	Command string
}

type Config struct {
	Environments []Environment `json:"environments"`
}

// checks if the provided runtimes are supported
func validateRuntimes(runtimes []string) error {
	for _, r := range runtimes {
		if !utils.Contains(r, supportedRuntimes) {
			return errors.New("invalid runtime " + r)
		}
	}
	return nil
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

	// validates runtimes
	var runtimes []string
	for _, env := range c.Environments {
		runtimes = append(runtimes, env.Runtime)
	}
	if err := validateRuntimes(runtimes); err != nil {
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
