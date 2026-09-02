package config

import (
	"testing"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/ratelimit"
	"github.com/glaciers-in-archives/snowman/internal/version"
)

func setupTest() SiteConfig {
	return CurrentSiteConfig
}

func teardownTest(original SiteConfig) {
	CurrentSiteConfig = original
}

type versionCompatibilityTest struct {
	name               string
	snowmanVersion     string
	currentVersion     version.Version
	expectedCompatible bool
}

var versionCompatibilityTests = []versionCompatibilityTest{
	// No version constraint - should always pass
	{
		name:               "No version constraint",
		snowmanVersion:     "",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: true,
	},

	// Compatible versions with >= constraint
	{
		name:               "Compatible with >= constraint (exact match)",
		snowmanVersion:     ">=0.7.1",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Compatible with >= constraint (higher version)",
		snowmanVersion:     ">=0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Compatible with >= constraint (development version)",
		snowmanVersion:     ">=0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: true,
	},

	// Incompatible versions with >= constraint
	{
		name:               "Incompatible with >= constraint",
		snowmanVersion:     ">=0.8.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: false,
	},

	// Caret constraint (^) - allows changes that don't modify left-most non-zero digit
	{
		name:               "Compatible with caret constraint",
		snowmanVersion:     "^0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Compatible with caret constraint (development)",
		snowmanVersion:     "^0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: true,
	},
	{
		name:               "Incompatible with caret constraint (major version)",
		snowmanVersion:     "^1.0.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},

	// Tilde constraint (~) - allows patch-level changes
	{
		name:               "Compatible with tilde constraint",
		snowmanVersion:     "~0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Compatible with tilde constraint (development)",
		snowmanVersion:     "~0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: true,
	},
	{
		name:               "Incompatible with tilde constraint (minor version)",
		snowmanVersion:     "~0.8.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},

	// Exact version match
	{
		name:               "Exact version match",
		snowmanVersion:     "0.7.1",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Exact version match (with development suffix)",
		snowmanVersion:     "0.7.1",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: true,
	},
	{
		name:               "Exact version mismatch",
		snowmanVersion:     "0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},

	// Range constraints
	{
		name:               "Compatible with range constraint",
		snowmanVersion:     ">=0.7.0 <0.8.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Compatible with range constraint (development)",
		snowmanVersion:     ">=0.7.0 <0.8.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "development"},
		expectedCompatible: true,
	},
	{
		name:               "Incompatible with range constraint (too high)",
		snowmanVersion:     ">=0.5.0 <0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},
	{
		name:               "Incompatible with range constraint (too low)",
		snowmanVersion:     ">=0.8.0 <1.0.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},

	// Invalid constraints
	{
		name:               "Invalid constraint format",
		snowmanVersion:     "not-a-valid-constraint",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},
	{
		name:               "Malformed constraint",
		snowmanVersion:     ">=0.7.0 <",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: ""},
		expectedCompatible: false,
	},

	// Edge cases with prerelease versions
	{
		name:               "Prerelease version with >= constraint",
		snowmanVersion:     ">=0.7.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 0, Suffix: "beta1"},
		expectedCompatible: true,
	},
	{
		name:               "Higher prerelease version",
		snowmanVersion:     ">=0.6.0",
		currentVersion:     version.Version{Major: 0, Minor: 7, Patch: 1, Suffix: "alpha"},
		expectedCompatible: true,
	},

	// Testing with different major versions
	{
		name:               "Major version 1.x compatible",
		snowmanVersion:     ">=1.0.0",
		currentVersion:     version.Version{Major: 1, Minor: 2, Patch: 3, Suffix: ""},
		expectedCompatible: true,
	},
	{
		name:               "Major version 1.x incompatible",
		snowmanVersion:     ">=2.0.0",
		currentVersion:     version.Version{Major: 1, Minor: 2, Patch: 3, Suffix: ""},
		expectedCompatible: false,
	},
}

func TestCheckVersionCompatibility(t *testing.T) {
	for _, test := range versionCompatibilityTests {
		t.Run(test.name, func(t *testing.T) {
			original := setupTest()
			defer teardownTest(original)

			CurrentSiteConfig.SnowmanVersion = test.snowmanVersion
			originalVersion := version.CurrentVersion
			version.CurrentVersion = test.currentVersion
			defer func() { version.CurrentVersion = originalVersion }()

			result := CheckVersionCompatibility()

			if result != test.expectedCompatible {
				t.Errorf("CheckVersionCompatibility() for version constraint '%s' with current version '%s': got %v, want %v",
					test.snowmanVersion,
					test.currentVersion.String(),
					result,
					test.expectedCompatible)
			}
		})
	}
}

// Test that CheckVersionCompatibility handles empty config correctly
func TestCheckVersionCompatibilityEmptyConfig(t *testing.T) {
	original := setupTest()
	defer teardownTest(original)

	CurrentSiteConfig = SiteConfig{}

	result := CheckVersionCompatibility()
	if !result {
		t.Error("CheckVersionCompatibility() with empty config should return true, got false")
	}
}

func TestRateLimitDefaults(t *testing.T) {
	var config SiteConfig
	if err := config.Parse([]byte("sparql_client:\n  endpoint: \"https://example.org/sparql\"\n")); err != nil {
		t.Fatal(err)
	}

	options := config.RateLimitOptions()
	defaults := ratelimit.DefaultConfig()

	if options != defaults {
		t.Errorf("Expected the default rate limit options %+v, got %+v", defaults, options)
	}
}

func TestRateLimitConfiguration(t *testing.T) {
	var config SiteConfig
	data := []byte(`
sparql_client:
  endpoint: "https://example.org/sparql"
rate_limit:
  max_retries: 10
  initial_backoff: "500ms"
  max_wait: "2m"
`)

	if err := config.Parse(data); err != nil {
		t.Fatal(err)
	}

	options := config.RateLimitOptions()
	if options.MaxRetries != 10 {
		t.Errorf("Expected 10 retries, got %d", options.MaxRetries)
	}
	if options.InitialBackoff != 500*time.Millisecond {
		t.Errorf("Expected an initial backoff of 500ms, got %v", options.InitialBackoff)
	}
	if options.MaxWait != 2*time.Minute {
		t.Errorf("Expected a max wait of 2m, got %v", options.MaxWait)
	}
}

func TestRateLimitCanBeDisabled(t *testing.T) {
	var config SiteConfig
	data := []byte(`
sparql_client:
  endpoint: "https://example.org/sparql"
rate_limit:
  max_retries: 0
`)

	if err := config.Parse(data); err != nil {
		t.Fatal(err)
	}

	if retries := config.RateLimitOptions().MaxRetries; retries != 0 {
		t.Errorf("Expected retrying to be disabled, got %d retries", retries)
	}
}

func TestInvalidRateLimitConfiguration(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "negative retries",
			data: "sparql_client:\n  endpoint: \"https://example.org/sparql\"\nrate_limit:\n  max_retries: -1\n",
		},
		{
			name: "unparsable duration",
			data: "sparql_client:\n  endpoint: \"https://example.org/sparql\"\nrate_limit:\n  initial_backoff: \"soon\"\n",
		},
		{
			name: "negative duration",
			data: "sparql_client:\n  endpoint: \"https://example.org/sparql\"\nrate_limit:\n  max_wait: \"-5s\"\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var config SiteConfig
			if err := config.Parse([]byte(test.data)); err == nil {
				t.Error("Expected the configuration to be rejected, got no error")
			}
		})
	}
}
