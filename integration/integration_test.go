package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/glaciers-in-archives/snowman/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"snowman": runSnowman,
	}))
}

// runSnowman runs the snowman CLI in-process for testing
func runSnowman() int {
	// Reset the command for each invocation
	// This is important because cobra commands maintain state

	// Note: We're calling cmd.Execute() which will use os.Args
	// The testscript framework sets up os.Args appropriately
	cmd.Execute()
	return 0
}

func TestIntegration(t *testing.T) {
	// Start a mock SPARQL server that will be used across all tests
	mockServer := NewMockSPARQLServer()
	t.Cleanup(mockServer.Close)

	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			// Set the mock server URL as an environment variable
			env.Setenv("MOCK_SPARQL_ENDPOINT", mockServer.URL)

			// Read and update snowman.yaml to use the mock server
			workDir := env.Getenv("WORK")
			configPath := workDir + "/snowman.yaml"

			if content, err := os.ReadFile(configPath); err == nil {
				updated := strings.ReplaceAll(string(content), "MOCK_SPARQL_ENDPOINT", mockServer.URL)
				if err := os.WriteFile(configPath, []byte(updated), 0644); err != nil {
					return err
				}
			}

			return nil
		},
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			// Custom commands can be added here if needed
			// For now, snowman is handled by TestMain
		},
	})
}
