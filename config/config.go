package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

// representation of json config file
type Config struct {
	Runtimes []string `json:runtimes`
	Machines []string `json:machines`
}

var supportedRuntimes = []string{
	"go",
	"node",
	"ruby",
	"clojure",
	"python",
}

var hyperSizes = []string{
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
}

func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func validateRuntime() error {
	return nil
}

// checks if provided machine size is on list of supported sizes
func validateSizes(sizes []string) error {
	for _, s := range sizes {
		if !contains(s, hyperSizes) {
			return errors.New("invalid machine size: " + s)
		}
	}
	return nil
}
func ValidateConfig(c *Config) error {

	if err := validateSizes(c.Machines); err != nil {
		return err
	}
	return nil
}

// parses json bytes into a Config struct
func ParseConfig(b []byte) (*Config, error) {
	c := &Config{}

	if err := json.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "invalid json")
	}

	if err := ValidateConfig(c); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	return c, nil
}

// reads the file at `path`
func ReadConfig() (*Config, error) {
	b, err := ioutil.ReadFile("ben.json")
	if err != nil {
		return nil, errors.Wrap(err, "file open failed")
	}
	return ParseConfig(b)
}
