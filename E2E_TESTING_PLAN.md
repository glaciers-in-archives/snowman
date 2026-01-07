# Integration and E2E Testing Plan for Snowman

## Executive Summary

This plan outlines the strategy for adding comprehensive integration and end-to-end (E2E) testing to Snowman, a static site generator for SPARQL endpoints. The plan is based on research of popular Go CLI tools (Hugo, GitHub CLI) and modern Go testing best practices.

## Current State

### Existing Test Coverage
- **6 test files** covering unit tests for:
  - Version compatibility checking (`config_test.go`)
  - SPARQL JSON parsing (`sparql_test.go`)
  - RDF term types (`rdf_test.go`)
  - Path validation (`utils_test.go`)
  - Math template functions (`math_test.go`)
  - String template functions (`strings_test.go`)

### Coverage Gaps
- ❌ No CLI command testing (build, server, new, cache)
- ❌ No integration tests with SPARQL endpoints
- ❌ No template rendering tests
- ❌ No view discovery tests
- ❌ No cache management tests
- ❌ No end-to-end workflow tests
- ❌ No fixture or test data infrastructure

## Testing Strategy

### Test Pyramid
```
        /\
       /E2E\      ← Few critical path tests (15 min)
      /------\
     /  INT   \   ← Component interaction tests (5 min)
    /----------\
   /   UNIT     \ ← Fast unit tests (2 min)
  /--------------\
```

### Three Testing Layers

#### 1. Unit Tests (Existing + Expand)
- **Current:** Template functions, config parsing, RDF/SPARQL processing
- **Expand:** Add more template function tests, edge cases
- **Run time:** < 2 minutes
- **Framework:** Go standard `testing` package + `testify/require`

#### 2. Integration Tests (New - Priority)
- **Focus:** CLI commands, file generation, template rendering
- **Run time:** < 5 minutes
- **Framework:** `testscript` (txtar-based, Hugo's approach)
- **Alternative:** `go-cmdtest` (simpler, Google's approach)

#### 3. E2E Tests (New)
- **Focus:** Complete user workflows, real SPARQL endpoints
- **Run time:** < 15 minutes
- **Framework:** Custom test harness with subprocess execution

## Recommended Approach: testscript + txtar

### Why testscript?

1. **Proven by Hugo** - Nearly identical use case (static site generator)
2. **Excellent for file generation** - Perfect for testing Snowman's output
3. **Txtar format** - Define entire site structures in readable text format
4. **Built-in golden file support** - Compare expected vs actual output
5. **Coverage support** - Integrates with `go test -cover`
6. **Go team approved** - Extracted from Go toolchain's own tests

### Txtar Format Example

```txtar
# Test building a simple site
-- snowman.yaml --
sparql_client:
  endpoint: "http://localhost:8080/sparql"
snowman_version: ">=0.7.0"

-- views.yaml --
views:
  - output: "index.html"
    query: "index.rq"
    template: "index.html"

-- queries/index.rq --
SELECT * WHERE { ?s ?p ?o } LIMIT 10

-- templates/layouts/base.html --
<!DOCTYPE html>
<html>{{ block "content" . }}{{ end }}</html>

-- templates/index.html --
{{ define "content" }}
<h1>Test Site</h1>
{{ end }}

-- expected/site/index.html --
<!DOCTYPE html>
<html>
<h1>Test Site</h1>
</html>
```

## Implementation Plan

### Phase 1: Setup Test Infrastructure (Week 1)

#### Task 1.1: Add Testing Dependencies
```bash
go get github.com/rogpeppe/go-internal/testscript
go get github.com/stretchr/testify
go get gotest.tools/v3/golden
```

#### Task 1.2: Create Directory Structure
```
snowman/
├── integration/
│   ├── integration_test.go       # Main test runner
│   ├── testdata/
│   │   ├── build_simple.txtar
│   │   ├── build_multipage.txtar
│   │   ├── cache_operations.txtar
│   │   └── error_cases.txtar
│   ├── golden/                   # Expected outputs
│   └── mock_sparql.go           # Mock SPARQL server
├── e2e/
│   ├── e2e_test.go              # End-to-end tests
│   ├── fixtures/                # Complete test sites
│   │   ├── wikidata_example/
│   │   └── complex_site/
│   └── helpers.go               # Test utilities
└── testdata/                     # Shared fixtures
```

#### Task 1.3: Create Mock SPARQL Server
Build a simple HTTP server that returns canned SPARQL JSON responses for testing without external dependencies.

### Phase 2: Integration Tests - Core Commands (Week 2)

#### Test Group 2.1: `snowman build` Command

**Test cases:**
- ✅ Build simple single-page site
- ✅ Build multi-page site with variables
- ✅ Build with SPARQL query execution
- ✅ Build with cache enabled (available strategy)
- ✅ Build with cache disabled (never strategy)
- ✅ Build static files only (`--static`)
- ✅ Build with custom config file (`--config`)
- ✅ Build with verbose output (`--verbose`)
- ✅ Build with timing info (`--timeit`)
- ❌ Build with missing template (error case)
- ❌ Build with invalid SPARQL query (error case)
- ❌ Build with invalid config (error case)

**Implementation:**
```go
// integration/build_test.go
func TestBuild(t *testing.T) {
    testscript.Run(t, testscript.Params{
        Dir: "testdata",
        Cmds: map[string]func(*testscript.TestScript, bool, []string){
            "snowman": cmdSnowman,
        },
    })
}
```

#### Test Group 2.2: `snowman new` Command

**Test cases:**
- ✅ Create new project with default name
- ✅ Create new project with custom directory
- ✅ Verify scaffold structure (files and directories)
- ✅ Verify scaffold file contents
- ❌ Create in existing directory (error case)

#### Test Group 2.3: `snowman cache` Commands

**Test cases:**
- ✅ Inspect empty SPARQL cache
- ✅ Inspect SPARQL cache after build
- ✅ Inspect specific query cache
- ✅ Clear all cache items
- ✅ Clear unused cache items
- ✅ Clear specific query cache
- ✅ Inspect resource cache
- ✅ Clear resource cache
- ✅ Verify cache directory structure

#### Test Group 2.4: `snowman server` Command

**Test cases:**
- ✅ Start server on default port (8000)
- ✅ Start server on custom port
- ✅ Start server on custom address
- ✅ Serve built site files
- ✅ Serve static assets (CSS, images)
- ❌ Start server on invalid port (error case)

### Phase 3: Integration Tests - Template Features (Week 3)

#### Test Group 3.1: Template Rendering

**Test cases:**
- ✅ Render with layouts (base.html)
- ✅ Render with includes
- ✅ Render with SPARQL query results
- ✅ Render multipage views
- ✅ Render with language-tagged literals
- ✅ Render with typed literals
- ✅ Render with unsafe templates (text mode)

#### Test Group 3.2: Template Functions

**Query Functions:**
- ✅ `query` function with SELECT
- ✅ `query_construct` function
- ✅ `query` with parameterized queries
- ✅ Query caching behavior

**Remote Functions:**
- ✅ `get_remote` function
- ✅ `get_remote_with_config` function
- ✅ Resource caching for remote fetches

**File Functions:**
- ✅ `read_file` function
- ✅ Include templates
- ✅ Include text files

**Config Functions:**
- ✅ `config` function (access metadata)
- ✅ `version` function
- ✅ `env` function

**Utility Functions:**
- ✅ `safe_html` function
- ✅ `uri` function
- ✅ `current_view` function

#### Test Group 3.3: View Discovery

**Test cases:**
- ✅ Discover views from views.yaml
- ✅ Parse static views
- ✅ Parse multipage views
- ✅ Extract multipage variables
- ❌ Invalid views.yaml syntax (error case)
- ❌ Missing query file (error case)
- ❌ Missing template file (error case)

### Phase 4: Integration Tests - Cache System (Week 4)

#### Test Group 4.1: SPARQL Cache

**Test cases:**
- ✅ Create cache on first build
- ✅ Use cache on second build (available strategy)
- ✅ Skip cache (never strategy)
- ✅ Cache parameterized queries
- ✅ Detect unused cache items
- ✅ Verify cache hash generation
- ✅ Verify last_build_queries.txt

#### Test Group 4.2: Resource Cache

**Test cases:**
- ✅ Cache remote JSON fetches
- ✅ Use cached resources (available strategy)
- ✅ Skip resource cache (never strategy)
- ✅ Verify cache directory structure

### Phase 5: E2E Tests - Complete Workflows (Week 5)

#### Test Group 5.1: New Project Workflow

**Scenario:** User creates and builds a new site
```bash
snowman new --directory="my-test-site"
cd my-test-site
# Modify files
snowman build
snowman server --port=8080
```

#### Test Group 5.2: Wikidata Example

**Scenario:** Build example site using real Wikidata endpoint
- Clone example project
- Run build with real SPARQL endpoint
- Verify output structure
- Verify data in generated pages

#### Test Group 5.3: Complex Multi-Page Site

**Scenario:** Site with multiple views, layouts, includes
- Multiple page types (index, list, detail)
- Shared layouts and includes
- SPARQL queries with various complexity
- Static assets (CSS, images, JS)

#### Test Group 5.4: Cache Workflow

**Scenario:** Build, cache, rebuild, clear cache
```bash
snowman build --cache-sparql=available
snowman cache sparql inspect
snowman build --cache-sparql=available  # Should use cache
snowman cache sparql clear --unused
snowman build --cache-sparql=never      # Should skip cache
```

### Phase 6: Error Cases and Edge Cases (Week 6)

#### Test Group 6.1: Configuration Errors

- ❌ Missing snowman.yaml
- ❌ Invalid YAML syntax
- ❌ Invalid version constraint
- ❌ Incompatible Snowman version
- ❌ Missing SPARQL endpoint

#### Test Group 6.2: SPARQL Errors

- ❌ SPARQL endpoint unreachable
- ❌ Invalid SPARQL syntax
- ❌ SPARQL endpoint returns error
- ❌ Malformed SPARQL JSON response

#### Test Group 6.3: Template Errors

- ❌ Missing template file
- ❌ Template syntax error
- ❌ Missing layout file
- ❌ Undefined template function
- ❌ Invalid function arguments

#### Test Group 6.4: File System Errors

- ❌ No write permission for output directory
- ❌ Disk full
- ❌ Invalid file paths
- ❌ Missing static directory

## Testing Best Practices

### 1. Golden Files

Use golden files for comparing expected output:

```go
// Update golden files
go test ./... -update

// Review changes
git diff integration/golden/

// Commit updated golden files
git commit -m "Update golden files for new feature"
```

### 2. Table-Driven Tests

Use Go's idiomatic table-driven pattern:

```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {name: "valid case", input: "test", expected: "TEST", wantErr: false},
    {name: "error case", input: "", expected: "", wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()  // Enable parallel execution
        got, err := function(tt.input)
        if tt.wantErr {
            require.Error(t, err)
            return
        }
        require.NoError(t, err)
        require.Equal(t, tt.expected, got)
    })
}
```

### 3. Use `require` over `assert`

**99% of the time, use `require`:**
- `require` stops test on failure (correct behavior)
- `assert` continues (usually not what you want)

```go
// Good
require.NoError(t, err)
require.Equal(t, expected, got)

// Avoid
assert.NoError(t, err)  // Continues even if error
assert.Equal(t, expected, got)  // May panic if got is nil
```

### 4. Parallel Test Execution

Enable parallel execution for faster test runs:

```go
func TestSomething(t *testing.T) {
    tests := []struct{ name string }{ /* ... */ }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Run subtests in parallel
            // ... test code
        })
    }
}
```

Can reduce test time by 40% in large projects.

### 5. Use t.TempDir()

Always use `t.TempDir()` for temporary directories:

```go
func TestBuild(t *testing.T) {
    dir := t.TempDir()  // Automatically cleaned up
    // ... use dir for test
}
```

### 6. testdata/ Directory

Place test fixtures in `testdata/` directory:
- Go tool ignores this directory
- Community standard (including stdlib)
- Clear separation from production code

## CI/CD Integration

### Test Execution Strategy

**On every commit:**
```bash
go test ./...  # Unit tests only (< 2 min)
```

**On pull request:**
```bash
go test ./...           # All unit tests
go test ./integration  # Integration tests (< 5 min)
```

**Before merge / nightly:**
```bash
go test ./...           # All tests
go test ./integration   # Integration tests
go test ./e2e          # E2E tests (< 15 min)
go test -race ./...    # Race detector
go test -cover ./...   # Coverage report
```

### GitHub Actions Workflow

```yaml
name: Tests
on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v ./...

  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v ./integration

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v ./e2e
```

## Success Metrics

### Code Coverage Targets
- **Unit tests:** 80% coverage of internal packages
- **Integration tests:** 70% coverage of cmd package
- **E2E tests:** Critical user paths (5-10 scenarios)

### Performance Targets
- **Unit tests:** < 2 minutes
- **Integration tests:** < 5 minutes
- **E2E tests:** < 15 minutes
- **Total test suite:** < 10 minutes (without E2E)

### Quality Metrics
- No flaky tests (tests must be deterministic)
- All tests pass consistently
- Golden files kept up to date
- Clear test names and documentation

## Dependencies to Add

```go
// go.mod
require (
    github.com/rogpeppe/go-internal v1.13.0     // testscript
    github.com/stretchr/testify v1.10.0         // assertions
    gotest.tools/v3 v3.6.0                       // golden files
)
```

## Documentation

### Test README

Create `integration/README.md`:
- How to run tests
- How to update golden files
- How to add new test cases
- How to debug failing tests

### Developer Guide

Add to main README.md:
```markdown
## Testing

### Run all tests
go test ./...

### Run integration tests
go test ./integration

### Run E2E tests
go test ./e2e

### Update golden files
go test ./... -update

### Run with coverage
go test -cover ./...

### Run with race detector
go test -race ./...
```

## Alternative Approach: go-cmdtest

If testscript is too complex, consider `go-cmdtest`:

**Pros:**
- Simpler learning curve
- Cross-platform shell-like language
- Good for straightforward CLI testing
- Easy golden file workflow

**Cons:**
- Less feature-rich than testscript
- No built-in coverage support
- Less proven for complex file generation testing

**Example:**
```bash
# build_simple.ct
$ snowman new --directory=test-site
Created new project: test-site

$ cd test-site
$ snowman build
Building site...
Generated 1 page

$ ls site/
index.html
```

## Next Steps

1. **Review this plan** with the team
2. **Choose testing framework** (testscript recommended)
3. **Start with Phase 1** (infrastructure setup)
4. **Implement Phase 2** (core commands)
5. **Iterate** through remaining phases
6. **Set up CI/CD** integration
7. **Document** testing approach

## Conclusion

This plan provides a comprehensive approach to adding integration and E2E testing to Snowman, based on proven patterns from Hugo, GitHub CLI, and the Go toolchain. The use of testscript with txtar format is particularly well-suited for testing a static site generator, providing clear, maintainable tests that verify both the CLI behavior and the generated output.

The phased approach allows for incremental progress, with each phase building on the previous one. Starting with test infrastructure and core commands ensures a solid foundation before tackling more complex scenarios.

**Estimated total effort:** 6 weeks
**Primary deliverable:** Comprehensive test suite covering CLI commands, template rendering, and end-to-end workflows
**Key benefit:** Confidence in making changes and adding features without regressions
