package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/glaciers-in-archives/snowman/internal/version"
	"gopkg.in/yaml.v2"
)

var CurrentSiteConfig SiteConfig

type ClientConfig struct {
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"http_headers"`
	// RateLimit is an optional minimum interval between SPARQL requests,
	// expressed as a Go duration string (e.g. "500ms", "1s").
	RateLimit string `yaml:"rate_limit"`
}

type SiteConfig struct {
	Client ClientConfig `yaml:"sparql_client"`
	// HTTPRateLimits maps a host to an optional minimum interval between
	// non-SPARQL HTTP requests to that host, expressed as a Go duration
	// string (e.g. "500ms", "1s").
	HTTPRateLimits map[string]string `yaml:"http_rate_limits"`
	SnowmanVersion string            `yaml:"snowman_version,omitempty"`
	Metadata       map[string]interface{}
}

func (c *SiteConfig) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	_, err := url.ParseRequestURI(c.Client.Endpoint) // #TODO why is https://example valid?
	if err != nil {
		return err
	}

	if c.Client.RateLimit != "" {
		if _, err := time.ParseDuration(c.Client.RateLimit); err != nil {
			return fmt.Errorf("invalid sparql_client.rate_limit %q: %w", c.Client.RateLimit, err)
		}
	}

	for host, value := range c.HTTPRateLimits {
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid http_rate_limits entry for %q: %w", host, err)
		}
	}

	return nil
}

// SparqlRateLimit returns the configured minimum interval between SPARQL
// requests, or zero if no rate limit is configured.
func (c *SiteConfig) SparqlRateLimit() time.Duration {
	d, _ := time.ParseDuration(c.Client.RateLimit)
	return d
}

// HTTPRateLimitIntervals returns the per-host minimum intervals between
// non-SPARQL HTTP requests.
func (c *SiteConfig) HTTPRateLimitIntervals() map[string]time.Duration {
	intervals := make(map[string]time.Duration, len(c.HTTPRateLimits))
	for host, value := range c.HTTPRateLimits {
		d, _ := time.ParseDuration(value)
		intervals[host] = d
	}
	return intervals
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
