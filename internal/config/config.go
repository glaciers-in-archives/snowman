package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/glaciers-in-archives/snowman/internal/ratelimit"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/glaciers-in-archives/snowman/internal/version"
	"gopkg.in/yaml.v2"
)

var CurrentSiteConfig SiteConfig

type ClientConfig struct {
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"http_headers"`
}

// RateLimitConfig describes how Snowman should behave when a SPARQL endpoint
// or a remote resource rate limits us. Durations are given as Go duration
// strings such as "500ms", "5s" or "2m".
type RateLimitConfig struct {
	MaxRetries     *int   `yaml:"max_retries,omitempty"`
	InitialBackoff string `yaml:"initial_backoff,omitempty"`
	MaxWait        string `yaml:"max_wait,omitempty"`
}

// Options returns the rate limiting options to use, filling in the defaults
// for the values the project doesn't configure.
func (c RateLimitConfig) Options() (ratelimit.Config, error) {
	options := ratelimit.DefaultConfig()

	if c.MaxRetries != nil {
		if *c.MaxRetries < 0 {
			return options, errors.New("rate_limit.max_retries can't be negative.")
		}
		options.MaxRetries = *c.MaxRetries
	}

	durations := []struct {
		name   string
		value  string
		target *time.Duration
	}{
		{"rate_limit.initial_backoff", c.InitialBackoff, &options.InitialBackoff},
		{"rate_limit.max_wait", c.MaxWait, &options.MaxWait},
	}

	for _, duration := range durations {
		if duration.value == "" {
			continue
		}

		parsed, err := time.ParseDuration(duration.value)
		if err != nil {
			return options, errors.New("Failed to parse " + duration.name + ", expected a duration such as \"5s\". Got: " + duration.value)
		}
		if parsed < 0 {
			return options, errors.New(duration.name + " can't be negative.")
		}
		*duration.target = parsed
	}

	return options, nil
}

type SiteConfig struct {
	Client         ClientConfig    `yaml:"sparql_client"`
	SnowmanVersion string          `yaml:"snowman_version,omitempty"`
	RateLimit      RateLimitConfig `yaml:"rate_limit,omitempty"`
	Metadata       map[string]interface{}
}

// RateLimitOptions returns the rate limiting options of the current
// configuration with the defaults applied.
func (c *SiteConfig) RateLimitOptions() ratelimit.Config {
	options, err := c.RateLimit.Options()
	if err != nil {
		// Invalid values are rejected while parsing the configuration.
		return ratelimit.DefaultConfig()
	}
	return options
}

func (c *SiteConfig) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	_, err := url.ParseRequestURI(c.Client.Endpoint) // #TODO why is https://example valid?
	if err != nil {
		return err
	}

	if _, err := c.RateLimit.Options(); err != nil {
		return err
	}

	return nil
}

func LoadConfig(fileLocation string) error {
	if _, err := os.Stat(fileLocation); err != nil {
		if fileLocation == "snowman.yaml" {
			return utils.ErrorExit("Unable to locate snowman.yaml in the current working directory.", err)
		} else {
			return utils.ErrorExit("Unable to locate Snowman configuration file at "+fileLocation+".", err)
		}
	}

	data, err := os.ReadFile(fileLocation)
	if err != nil {
		return utils.ErrorExit("Failed to read "+fileLocation+".", err)
	}

	if err := CurrentSiteConfig.Parse(data); err != nil {
		return utils.ErrorExit("Failed to parse "+fileLocation+".", err)
	}

	return nil
}

// CheckVersionCompatibility checks if the current Snowman version satisfies
// the version constraint specified in snowman.yaml (if any).
// Returns true if compatible or no constraint specified, false otherwise.
func CheckVersionCompatibility() bool {
	if CurrentSiteConfig.SnowmanVersion == "" {
		return true
	}

	constraint, err := semver.NewConstraint(CurrentSiteConfig.SnowmanVersion)
	if err != nil {
		return false
	}

	// Get current version string (e.g., "0.7.1-development")
	currentVersionStr := version.CurrentVersion.String()

	currentVer, err := semver.NewVersion(currentVersionStr)
	if err != nil {
		return false
	}

	// Check if current version satisfies the constraint
	// For prerelease/development versions, also check the core version without prerelease suffix
	// This allows 0.7.1-development to satisfy constraints like >=0.7.0
	satisfiesConstraint := constraint.Check(currentVer)

	if !satisfiesConstraint && currentVer.Prerelease() != "" {
		// Try checking with just the core version (without prerelease)
		coreVersionStr := fmt.Sprintf("%d.%d.%d", currentVer.Major(), currentVer.Minor(), currentVer.Patch())
		coreVer, err := semver.NewVersion(coreVersionStr)
		if err == nil {
			satisfiesConstraint = constraint.Check(coreVer)
		}
	}

	return satisfiesConstraint
}
