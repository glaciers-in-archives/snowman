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
}

// URLRule holds the settings applied to the HTTP requests Snowman makes to the
// URLs matching a pattern, be it the SPARQL endpoint or a remote resource.
type URLRule struct {
	// RateLimit is a minimum interval between the matching requests,
	// expressed as a Go duration string (e.g. "500ms", "1s").
	RateLimit string `yaml:"rate_limit"`
}

type SiteConfig struct {
	Client ClientConfig `yaml:"sparql_client"`
	// URLRules maps a URL pattern, such as "example.org" or
	// "example.org/api/*", to the settings used for the requests matching it.
	URLRules       map[string]URLRule `yaml:"url_rules"`
	SnowmanVersion string             `yaml:"snowman_version,omitempty"`
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

	for pattern, rule := range c.URLRules {
		if rule.RateLimit == "" {
			continue
		}
		if _, err := time.ParseDuration(rule.RateLimit); err != nil {
			return fmt.Errorf("invalid rate_limit for url_rules entry %q: %w", pattern, err)
		}
	}

	return nil
}

// RateLimitIntervals returns the minimum intervals between the HTTP requests
// Snowman makes to the URLs matching each pattern.
func (c *SiteConfig) RateLimitIntervals() map[string]time.Duration {
	intervals := make(map[string]time.Duration, len(c.URLRules))
	for pattern, rule := range c.URLRules {
		if rule.RateLimit == "" {
			continue
		}
		d, _ := time.ParseDuration(rule.RateLimit)
		intervals[pattern] = d
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
