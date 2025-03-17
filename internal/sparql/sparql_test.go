package sparql

import (
	"strings"
	"testing"

	"github.com/glaciers-in-archives/snowman/internal/rdf"
)

func TestParseSPARQLJSONWithLanguageTags(t *testing.T) {
	// Test case based on the provided example JSON file
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

	// Parse the JSON data
	reader := strings.NewReader(jsonData)
	results := ParseSPARQLJSON(reader)

	// Validate the results
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Check if the name field exists in the first result
	if term, exists := results[0]["name"]; !exists {
		t.Fatal("Expected 'name' field in result, but it wasn't found")
	} else {
		// Verify it's a Literal
		if term.Type() != rdf.TermLiteral {
			t.Fatalf("Expected 'name' to be a Literal, got %v", term.Type())
		}

		// Convert to Literal to access Lang method
		literal, ok := term.(rdf.Literal)
		if !ok {
			t.Fatal("Could not convert term to Literal")
		}

		// Check the value
		if literal.String() != "Canada" {
			t.Fatalf("Expected 'Canada', got '%s'", literal.String())
		}

		// Check the language tag
		if literal.Lang() != "en" {
			t.Fatalf("Expected language tag 'en', got '%s'", literal.Lang())
		}

		// Check the datatype is rdf:langString
		expectedDataType := "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"
		if literal.DataType.String() != expectedDataType {
			t.Fatalf("Expected datatype '%s', got '%s'", expectedDataType, literal.DataType.String())
		}
	}
}

// Test with multiple language tags
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

	// Parse the JSON data
	reader := strings.NewReader(jsonData)
	results := ParseSPARQLJSON(reader)

	// Validate the results
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Expected test data
	expectedData := []struct {
		value string
		lang  string
	}{
		{"Canada", "en"},
		{"Canada", "fr"},
		{"Kanada", "de"},
	}

	// Check each result
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

// Test mixed literals with and without language tags
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

	// Parse the JSON data
	reader := strings.NewReader(jsonData)
	results := ParseSPARQLJSON(reader)

	// Validate the results
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Check the name (with language tag)
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

	// Check the population (typed literal without language)
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

	// Check the description (plain literal without language or explicit datatype)
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

		// Plain literals should be typed as xsd:string
		expectedDataType := "http://www.w3.org/2001/XMLSchema#string"
		if literal.DataType.String() != expectedDataType {
			t.Fatalf("Expected datatype '%s', got '%s'", expectedDataType, literal.DataType.String())
		}
	}
}
