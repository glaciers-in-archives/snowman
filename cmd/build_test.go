package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

// sparqlResponse is what the test endpoint answers to every query: two rows with the ids "a" and "b".
const sparqlResponse = `{"head":{"vars":["id"]},"results":{"bindings":[
	{"id":{"type":"literal","value":"a"}},
	{"id":{"type":"literal","value":"b"}}
]}}`

// setupProject creates a Snowman project in a temporary directory, makes it the
// working directory and returns a counter of the queries its endpoint received.
func setupProject(t *testing.T) *int {
	t.Helper()

	queryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryCount++
		w.Header().Set("Content-Type", "application/sparql-results+json")
		fmt.Fprint(w, sparqlResponse)
	}))
	t.Cleanup(server.Close)

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(workingDirectory) })

	writeFile(t, "snowman.yaml", "sparql_client:\n  endpoint: \""+server.URL+"\"\n")
	writeFile(t, "views.yaml", `views:
  - output: "index.html"
    query: "entities.rq"
    template: "index.html"
  - output: "entities/{{ id }}.html"
    query: "entities.rq"
    template: "entity.html"
`)
	writeFile(t, "queries/entities.rq", "SELECT ?id WHERE { ?id ?p ?o }")
	writeFile(t, "queries/entity.rq", "SELECT ?id WHERE { <{{.}}> ?p ?id }")
	writeFile(t, "static/style.css", "body {}")
	writeFile(t, "templates/index.html", "index v1")
	writeFile(t, "templates/entity.html", `entity {{ .id }} v1{{ $details := query "entity.rq" .id }}`)

	return &queryCount
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0660); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expected {
		t.Errorf("Expected %s to contain %q, got %q", path, expected, string(content))
	}
}

func runBuild(args ...string) error {
	staticBuildOption = false
	rootCmd.SetArgs(append([]string{"build"}, args...))
	return rootCmd.Execute()
}

// updateTemplates changes all templates so that re-rendered pages can be told apart from old ones.
func updateTemplates(t *testing.T) {
	t.Helper()
	writeFile(t, "templates/index.html", "index v2")
	writeFile(t, "templates/entity.html", `entity {{ .id }} v2{{ $details := query "entity.rq" .id }}`)
}

func TestFullBuild(t *testing.T) {
	setupProject(t)
	writeFile(t, "site/leftover.html", "leftover")

	if err := runBuild(); err != nil {
		t.Fatal(err)
	}

	assertFile(t, "site/index.html", "index v1")
	assertFile(t, "site/entities/a.html", "entity a v1")
	assertFile(t, "site/entities/b.html", "entity b v1")
	assertFile(t, "site/style.css", "body {}")
	if _, err := os.Stat("site/leftover.html"); !os.IsNotExist(err) {
		t.Error("Expected a full build to empty the site directory.")
	}
}

func TestPartialBuildOfView(t *testing.T) {
	queryCount := setupProject(t)
	if err := runBuild(); err != nil {
		t.Fatal(err)
	}
	queriesAfterFullBuild := *queryCount

	updateTemplates(t)
	writeFile(t, "site/leftover.html", "leftover")
	if err := runBuild("entities/{{id}}.html"); err != nil {
		t.Fatal(err)
	}

	assertFile(t, "site/entities/a.html", "entity a v2")
	assertFile(t, "site/entities/b.html", "entity b v2")
	// everything else is left in place
	assertFile(t, "site/index.html", "index v1")
	assertFile(t, "site/style.css", "body {}")
	assertFile(t, "site/leftover.html", "leftover")

	if *queryCount != queriesAfterFullBuild {
		t.Errorf("Expected the partial build to use the cache, but it issued %d queries.", *queryCount-queriesAfterFullBuild)
	}
}

func TestPartialBuildOfSinglePageView(t *testing.T) {
	setupProject(t)
	if err := runBuild(); err != nil {
		t.Fatal(err)
	}

	updateTemplates(t)
	if err := runBuild("index.html"); err != nil {
		t.Fatal(err)
	}

	assertFile(t, "site/index.html", "index v2")
	assertFile(t, "site/entities/a.html", "entity a v1")
	assertFile(t, "site/entities/b.html", "entity b v1")
}

func TestPartialBuildOfViewWithArgument(t *testing.T) {
	setupProject(t)
	if err := runBuild(); err != nil {
		t.Fatal(err)
	}

	updateTemplates(t)
	// the whitespace around the variable doesn't need to match views.yaml
	if err := runBuild("entities/{{  id}}.html", "b"); err != nil {
		t.Fatal(err)
	}

	assertFile(t, "site/entities/b.html", "entity b v2")
	assertFile(t, "site/entities/a.html", "entity a v1")
	assertFile(t, "site/index.html", "index v1")
}

func TestPartialBuildWithoutExistingSite(t *testing.T) {
	setupProject(t)

	if err := runBuild("entities/{{id}}.html", "a"); err != nil {
		t.Fatal(err)
	}

	assertFile(t, "site/entities/a.html", "entity a v1")
	for _, path := range []string{"site/entities/b.html", "site/index.html", "site/style.css"} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected %s to not be built.", path)
		}
	}
}

func TestPartialBuildKeepsCacheHistory(t *testing.T) {
	setupProject(t)
	if err := runBuild(); err != nil {
		t.Fatal(err)
	}
	historyAfterFullBuild, err := utils.ReadLineSeperatedFile(".snowman/last_build_queries.txt")
	if err != nil {
		t.Fatal(err)
	}
	// entities.rq and entity.rq with the arguments a and b
	if len(historyAfterFullBuild) != 3 {
		t.Fatalf("Expected the full build to use 3 cache items, got %d", len(historyAfterFullBuild))
	}

	if err := runBuild("entities/{{id}}.html", "a"); err != nil {
		t.Fatal(err)
	}

	// the cache items of the pages which weren't rebuilt are still in use
	historyAfterPartialBuild, err := utils.ReadLineSeperatedFile(".snowman/last_build_queries.txt")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(historyAfterFullBuild)
	sort.Strings(historyAfterPartialBuild)
	if !reflect.DeepEqual(historyAfterFullBuild, historyAfterPartialBuild) {
		t.Errorf("Expected the partial build to keep the cache history %v, got %v", historyAfterFullBuild, historyAfterPartialBuild)
	}
}

func TestPartialBuildErrors(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{"unknown view", []string{"unknown.html"}, "Unable to find a view with the output unknown.html"},
		{"unknown argument", []string{"entities/{{id}}.html", "c"}, "did not return any result where id is c"},
		{"argument for single page view", []string{"index.html", "a"}, "doesn't take an argument"},
		{"empty view", []string{""}, "can't be empty"},
		{"combined with static", []string{"index.html", "--static"}, "can't be combined"},
		{"too many arguments", []string{"entities/{{id}}.html", "a", "b"}, "accepts at most 2 arg(s)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupProject(t)
			if err := runBuild(); err != nil {
				t.Fatal(err)
			}

			err := runBuild(tt.args...)
			if err == nil {
				t.Fatal("Expected an error.")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected an error containing %q, got %q", tt.expectedError, err.Error())
			}

			// a failed partial build doesn't touch the rest of the site
			assertFile(t, "site/index.html", "index v1")
			assertFile(t, "site/entities/a.html", "entity a v1")
			assertFile(t, "site/entities/b.html", "entity b v1")
		})
	}
}
