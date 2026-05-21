package sparql

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/rdf"
)

func TestParseSPARQLJSONWithLanguageTags(t *testing.T) {
	jsonData := `{
		"head": {
			"vars": ["name"]
		},
		"results": {
			"bindings": [{
				"name": {
					"xml:lang": "en",
					"type": "literal",
					"value": "Canada"
				}
			}]
		}
	}`

	reader := strings.NewReader(jsonData)
	results, err := ParseSPARQLJSON(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if term, exists := results[0]["name"]; !exists {
		t.Fatal("Expected 'name' field in result, but it wasn't found")
	} else {
		if term.Type() != rdf.TermLiteral {
			t.Fatalf("Expected 'name' to be a Literal, got %v", term.Type())
		}

		literal, ok := term.(rdf.Literal)
		if !ok {
			t.Fatal("Could not convert term to Literal")
		}

		if literal.String() != "Canada" {
			t.Fatalf("Expected 'Canada', got '%s'", literal.String())
		}

		if literal.Lang() != "en" {
			t.Fatalf("Expected language tag 'en', got '%s'", literal.Lang())
		}

		expectedDataType := "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"
		if literal.DataType.String() != expectedDataType {
			t.Fatalf("Expected datatype '%s', got '%s'", expectedDataType, literal.DataType.String())
		}
	}
}

func TestParseSPARQLJSONWithMultipleLanguageTags(t *testing.T) {
	jsonData := `{
		"head": {
			"vars": ["name"]
		},
		"results": {
			"bindings": [
				{
					"name": {
						"xml:lang": "en",
						"type": "literal",
						"value": "Canada"
					}
				},
				{
					"name": {
						"xml:lang": "fr",
						"type": "literal",
						"value": "Canada"
					}
				},
				{
					"name": {
						"xml:lang": "de",
						"type": "literal",
						"value": "Kanada"
					}
				}
			]
		}
	}`

	reader := strings.NewReader(jsonData)
	results, err := ParseSPARQLJSON(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	expectedData := []struct {
		value string
		lang  string
	}{
		{"Canada", "en"},
		{"Canada", "fr"},
		{"Kanada", "de"},
	}

	for i, expected := range expectedData {
		if term, exists := results[i]["name"]; !exists {
			t.Fatalf("Result %d: Expected 'name' field, but it wasn't found", i)
		} else {
			literal, ok := term.(rdf.Literal)
			if !ok {
				t.Fatalf("Result %d: Could not convert term to Literal", i)
			}

			if literal.String() != expected.value {
				t.Fatalf("Result %d: Expected value '%s', got '%s'", i, expected.value, literal.String())
			}

			if literal.Lang() != expected.lang {
				t.Fatalf("Result %d: Expected language tag '%s', got '%s'", i, expected.lang, literal.Lang())
			}
		}
	}
}

func TestParseSPARQLJSONWithMixedLiterals(t *testing.T) {
	jsonData := `{
		"head": {
			"vars": ["name", "population", "description"]
		},
		"results": {
			"bindings": [
				{
					"name": {
						"xml:lang": "en",
						"type": "literal",
						"value": "Canada"
					},
					"population": {
						"type": "literal",
						"datatype": "http://www.w3.org/2001/XMLSchema#integer",
						"value": "38000000"
					},
					"description": {
						"type": "literal",
						"value": "A country in North America"
					}
				}
			]
		}
	}`

	reader := strings.NewReader(jsonData)
	results, err := ParseSPARQLJSON(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if term, exists := results[0]["name"]; !exists {
		t.Fatal("Expected 'name' field, but it wasn't found")
	} else {
		literal, ok := term.(rdf.Literal)
		if !ok {
			t.Fatal("Could not convert 'name' term to Literal")
		}

		if literal.String() != "Canada" {
			t.Fatalf("Expected 'Canada', got '%s'", literal.String())
		}

		if literal.Lang() != "en" {
			t.Fatalf("Expected language tag 'en', got '%s'", literal.Lang())
		}
	}

	if term, exists := results[0]["population"]; !exists {
		t.Fatal("Expected 'population' field, but it wasn't found")
	} else {
		literal, ok := term.(rdf.Literal)
		if !ok {
			t.Fatal("Could not convert 'population' term to Literal")
		}

		if literal.String() != "38000000" {
			t.Fatalf("Expected '38000000', got '%s'", literal.String())
		}

		if literal.Lang() != "" {
			t.Fatalf("Expected empty language tag, got '%s'", literal.Lang())
		}

		expectedDataType := "http://www.w3.org/2001/XMLSchema#integer"
		if literal.DataType.String() != expectedDataType {
			t.Fatalf("Expected datatype '%s', got '%s'", expectedDataType, literal.DataType.String())
		}
	}

	if term, exists := results[0]["description"]; !exists {
		t.Fatal("Expected 'description' field, but it wasn't found")
	} else {
		literal, ok := term.(rdf.Literal)
		if !ok {
			t.Fatal("Could not convert 'description' term to Literal")
		}

		if literal.String() != "A country in North America" {
			t.Fatalf("Expected 'A country in North America', got '%s'", literal.String())
		}

		if literal.Lang() != "" {
			t.Fatalf("Expected empty language tag, got '%s'", literal.Lang())
		}

		expectedDataType := "http://www.w3.org/2001/XMLSchema#string"
		if literal.DataType.String() != expectedDataType {
			t.Fatalf("Expected datatype '%s', got '%s'", expectedDataType, literal.DataType.String())
		}
	}
}

// newTestRepository creates a Repository configured to hit a test HTTP server.
func newTestRepository(serverURL string, retry config.RetryConfig) Repository {
	return Repository{
		client: config.ClientConfig{
			Endpoint: serverURL,
			Retry:    retry,
		},
		httpClient:   http.DefaultClient,
		CacheManager: cache.SparqlCacheManager{},
	}
}

func TestQueryCallSucceeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"head":{"vars":[]},"results":{"bindings":[]}}`))
	}))
	defer srv.Close()

	repo := newTestRepository(srv.URL, config.RetryConfig{})
	result, err := repo.QueryCall("SELECT * WHERE {}", "application/sparql-results+json")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestQueryCallRetryOnServerError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"head":{"vars":[]},"results":{"bindings":[]}}`))
	}))
	defer srv.Close()

	repo := newTestRepository(srv.URL, config.RetryConfig{
		Enabled:        true,
		MaxRetries:     3,
		InitialDelayMS: 1,
	})

	result, err := repo.QueryCall("SELECT * WHERE {}", "application/sparql-results+json")
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if attempts != 3 {
		t.Fatalf("Expected 3 attempts, got %d", attempts)
	}
}

func TestQueryCallNoRetryWhenDisabled(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	repo := newTestRepository(srv.URL, config.RetryConfig{Enabled: false, MaxRetries: 3, InitialDelayMS: 1})

	_, err := repo.QueryCall("SELECT * WHERE {}", "application/sparql-results+json")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if attempts != 1 {
		t.Fatalf("Expected exactly 1 attempt when retry is disabled, got %d", attempts)
	}
}

func TestQueryCallCriticalOn429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	repo := newTestRepository(srv.URL, config.RetryConfig{
		Enabled:        true,
		MaxRetries:     3,
		InitialDelayMS: 1,
	})

	_, err := repo.QueryCall("SELECT * WHERE {}", "application/sparql-results+json")
	if err == nil {
		t.Fatal("Expected error for 429 response")
	}

	var criticalErr *CriticalEndpointError
	if !errors.As(err, &criticalErr) {
		t.Fatalf("Expected CriticalEndpointError for 429, got: %T %v", err, err)
	}
	if criticalErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429, got %d", criticalErr.StatusCode)
	}
	if attempts != 1 {
		t.Fatalf("Expected exactly 1 attempt for 429 (no retry), got %d", attempts)
	}
}

func TestQueryCallCriticalOn4xx(t *testing.T) {
	for _, code := range []int{400, 401, 403, 404} {
		code := code
		t.Run(http.StatusText(code), func(t *testing.T) {
			attempts := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				w.WriteHeader(code)
			}))
			defer srv.Close()

			repo := newTestRepository(srv.URL, config.RetryConfig{
				Enabled:        true,
				MaxRetries:     3,
				InitialDelayMS: 1,
			})

			_, err := repo.QueryCall("SELECT * WHERE {}", "application/sparql-results+json")
			if err == nil {
				t.Fatalf("Expected error for %d response", code)
			}

			var criticalErr *CriticalEndpointError
			if !errors.As(err, &criticalErr) {
				t.Fatalf("Expected CriticalEndpointError for %d, got: %T", code, err)
			}
			if attempts != 1 {
				t.Fatalf("Expected exactly 1 attempt for %d (no retry), got %d", code, attempts)
			}
		})
	}
}

func TestQueryCallExhaustsRetries(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	repo := newTestRepository(srv.URL, config.RetryConfig{
		Enabled:        true,
		MaxRetries:     2,
		InitialDelayMS: 1,
	})

	_, err := repo.QueryCall("SELECT * WHERE {}", "application/sparql-results+json")
	if err == nil {
		t.Fatal("Expected error after retries exhausted")
	}
	if attempts != 3 {
		t.Fatalf("Expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestIsCriticalStatusCode(t *testing.T) {
	critical := []int{400, 401, 403, 404, 429, 422, 499}
	for _, code := range critical {
		if !isCriticalStatusCode(code) {
			t.Errorf("Expected %d to be critical", code)
		}
	}

	retryable := []int{500, 502, 503, 504}
	for _, code := range retryable {
		if isCriticalStatusCode(code) {
			t.Errorf("Expected %d to be retryable (not critical)", code)
		}
	}
}
