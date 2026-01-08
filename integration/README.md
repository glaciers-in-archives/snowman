# Integration Testing - Proof of Concept

This directory contains a working proof of concept for integration testing of Snowman's `build` command using the testscript framework.

## ✅ What's Working

The POC successfully demonstrates:

1. **testscript + txtar format** - Clean, readable test files that define entire site structures
2. **Mock SPARQL server** - Isolated testing without external dependencies
3. **In-process testing** - Fast test execution using Snowman as a library
4. **Parallel test execution** - Tests run in parallel for speed
5. **Complete build verification** - Tests verify file generation, content, and structure

## 📊 Test Results

```bash
$ cd integration && go test -v
=== RUN   TestIntegration
=== RUN   TestIntegration/build_multipage
=== RUN   TestIntegration/build_simple
--- PASS: TestIntegration (0.00s)
    --- PASS: TestIntegration/build_simple (0.05s)
    --- PASS: TestIntegration/build_multipage (0.05s)
PASS
ok  	github.com/glaciers-in-archives/snowman/integration	0.068s
```

✅ **2 tests passing** in ~70ms

## 📁 Structure

```
integration/
├── README.md                    # This file
├── integration_test.go          # Main test runner
├── mock_sparql.go              # Mock SPARQL server
└── testdata/                    # Test cases in txtar format
    ├── build_simple.txtar       # Single-page site test
    └── build_multipage.txtar    # Multi-page site test
```

## 🧪 Test Cases

### 1. build_simple.txtar

Tests building a simple single-page site with:
- ✅ snowman.yaml configuration
- ✅ views.yaml with single static view
- ✅ SPARQL query file
- ✅ Template with layout inheritance
- ✅ Static file copying (CSS)
- ✅ Content verification in generated HTML
- ✅ Cache directory creation

**What it validates:**
- Build command executes successfully
- Output files are created in correct locations
- Templates render with SPARQL data
- Static files are copied
- Cache system is initialized

### 2. build_multipage.txtar

Tests building a multi-page site with:
- ✅ Multiple views (index + items)
- ✅ Multipage view with `{{id}}` variable
- ✅ Multiple HTML files generated from single template
- ✅ SPARQL query returning multiple results
- ✅ Template iteration with `{{ range . }}`

**What it validates:**
- Multipage view generation works correctly
- Multiple files created from one template
- Variable extraction from SPARQL results ({{id}})
- Content unique to each generated page

## 🚀 Running Tests

### Run all integration tests
```bash
cd integration
go test -v
```

### Run specific test
```bash
go test -v -run TestIntegration/build_simple
go test -v -run TestIntegration/build_multipage
```

### Run with coverage
```bash
go test -cover
```

### Run with race detector
```bash
go test -race
```

## 📝 txtar Format

The txtar format is a simple text-based archive format:

```txtar
# Comments start with #
# Commands are specified with test directives

# Execute a command
exec snowman build

# Check stdout contains text
stdout 'Building project'

# Verify file exists
exists site/index.html

# Check file contains text
grep 'Welcome' site/index.html

-- filename1 --
file contents here

-- subdir/filename2 --
more file contents
```

### Available Test Directives

- `exec <command>` - Execute command (must succeed)
- `! exec <command>` - Execute command (must fail)
- `stdout <pattern>` - Check stdout contains pattern
- `stderr <pattern>` - Check stderr contains pattern
- `exists <path>` - Verify file exists
- `! exists <path>` - Verify file doesn't exist
- `grep <pattern> <file>` - Check file contains pattern
- `! grep <pattern> <file>` - Check file doesn't contain pattern
- `cmp <file1> <file2>` - Compare two files

## 🔧 How It Works

### 1. Mock SPARQL Server

`mock_sparql.go` provides a test HTTP server that returns canned SPARQL JSON responses:

```go
mockServer := NewMockSPARQLServer()
defer mockServer.Close()
```

The mock server recognizes query patterns:
- Queries containing "index" → returns title and description
- Queries containing "items" → returns multiple items with id, name, description
- Other queries → returns empty results

### 2. Test Setup

For each test, the framework:

1. Creates an isolated sandbox directory (`$WORK`)
2. Extracts files from txtar archive into sandbox
3. Starts mock SPARQL server
4. Replaces `MOCK_SPARQL_ENDPOINT` in config files with real server URL
5. Runs test commands in sandbox
6. Cleans up after test

### 3. In-Process Execution

Snowman runs in-process via `cmd.Execute()`:

```go
func runSnowman() int {
    cmd.Execute()
    return 0
}
```

This is faster than spawning subprocesses and easier to debug.

## 🎯 Template Patterns

### Single-page views (list/collection)

Query returns multiple results → template iterates over them:

```go
{{ range . }}
  <li>{{ .field_name }}</li>
{{ end }}
```

### Multi-page views (detail pages)

Query returns multiple results → each result becomes a separate page:

```go
# views.yaml
- output: "items/{{id}}.html"
  query: "items.rq"
  template: "item.html"

# Template accesses binding fields directly:
<h1>{{ .name }}</h1>
<p>{{ .description }}</p>
```

### Layout inheritance

```go
# Layout (templates/layouts/base.html)
{{ define "base" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ block "title" . }}{{ end }}</title>
</head>
<body>
  {{ block "content" . }}{{ end }}
</body>
</html>
{{ end }}

# Page template (templates/index.html)
{{ template "base" . }}

{{ define "title" }}My Title{{ end }}

{{ define "content" }}
<h1>Content here</h1>
{{ end }}
```

## 📚 Adding New Tests

### 1. Create a new txtar file

```bash
touch integration/testdata/my_new_test.txtar
```

### 2. Define the test structure

```txtar
# Description of what this test validates

# Test commands
exec snowman build
stdout 'expected output'
exists expected/file.html

-- snowman.yaml --
sparql_client:
  endpoint: "MOCK_SPARQL_ENDPOINT"
snowman_version: ">=0.7.0"

-- views.yaml --
views:
  - output: "index.html"
    query: "index.rq"
    template: "index.html"

-- queries/index.rq --
SELECT ?title WHERE {
  VALUES (?placeholder) { ("index") }
}

-- templates/layouts/base.html --
{{ define "base" }}
<!DOCTYPE html>
<html>
  {{ block "content" . }}{{ end }}
</html>
{{ end }}

-- templates/index.html --
{{ template "base" . }}
{{ define "content" }}
  {{ range . }}
    <h1>{{ .title }}</h1>
  {{ end }}
{{ end }}
```

### 3. Run your test

```bash
go test -v -run TestIntegration/my_new_test
```

## 🔍 Debugging Failed Tests

When a test fails, testscript shows:
- The command that failed
- stdout and stderr output
- The working directory path
- Environment variables

Example:

```
FAIL: testdata/build_simple.txtar:20: no match for `Welcome` found in site/index.html
```

To debug:
1. Look at the line number in the txtar file
2. Check the stdout/stderr output in test results
3. Temporarily add `cat site/index.html` to see actual content
4. Check template syntax matches Snowman's expectations

## 🎨 Mock Server Customization

To add new mock responses, edit `mock_sparql.go`:

```go
case containsString(query, "my_keyword"):
    response = `{
      "head": {
        "vars": ["field1", "field2"]
      },
      "results": {
        "bindings": [
          {
            "field1": {"type": "literal", "value": "value1"},
            "field2": {"type": "literal", "value": "value2"}
          }
        ]
      }
    }`
```

Then use `"my_keyword"` in your test queries to trigger this response.

## ✨ Benefits of This Approach

1. **Fast** - Tests run in ~70ms, can run in parallel
2. **Isolated** - No external dependencies, no network calls
3. **Readable** - txtar format is clear and self-documenting
4. **Maintainable** - Easy to add new tests, update expectations
5. **Comprehensive** - Tests entire build pipeline, not just units
6. **Proven** - Same approach used by Hugo and Go toolchain
7. **Debuggable** - Clear error messages, easy to inspect test state

## 🔮 Next Steps

This POC demonstrates the approach works well. To implement the full testing plan:

1. **Add more build tests:**
   - Cache behavior (available/never strategies)
   - Error cases (missing files, invalid config)
   - Different template functions
   - Static-only builds (`--static` flag)
   - Verbose output (`--verbose` flag)

2. **Add other command tests:**
   - `snowman new` - Project scaffolding
   - `snowman cache` - Cache inspection and clearing
   - `snowman server` - Development server
   - `snowman version` - Version display

3. **Add template function tests:**
   - Query functions (`query`, `query_construct`)
   - Remote fetching (`get_remote`)
   - String functions (`split`, `join`, etc.)
   - Dict functions (`dict_create`, `dict_set`, etc.)

4. **Set up CI/CD:**
   - Add GitHub Actions workflow
   - Run tests on every commit
   - Report coverage

## 📖 References

- [testscript documentation](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
- [testscript article](https://bitfieldconsulting.com/posts/cli-testing)
- [Hugo integration tests](https://github.com/gohugoio/hugo/tree/master/hugolib) (similar approach)
- [Testing plan](../E2E_TESTING_PLAN.md)

## 🤝 Contributing

When adding new tests:
1. Follow the existing pattern in `testdata/`
2. Use descriptive test names
3. Add comments explaining what's being tested
4. Verify both success and failure cases
5. Keep tests focused and minimal
6. Run `go test -v` to verify before committing

---

**Status:** ✅ Proof of Concept Complete
**Tests Passing:** 2/2
**Coverage:** Build command with single-page and multi-page views
**Next:** Expand test coverage as outlined in E2E_TESTING_PLAN.md
