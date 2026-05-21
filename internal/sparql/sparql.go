package sparql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/rdf"
	"github.com/spf13/cast"
)

type Repository struct {
	client       config.ClientConfig
	httpClient   *http.Client
	verbose      bool
	CacheManager cache.SparqlCacheManager
	QueryIndex   map[string]string
}

var CurrentRepository Repository

func NewRepository(cacheManager cache.SparqlCacheManager, queryIndex map[string]string, verbose bool) error {
	repo := Repository{
		client:     config.CurrentSiteConfig.Client,
		QueryIndex: queryIndex,
		verbose:    verbose,
	}
	repo.httpClient = http.DefaultClient

	repo.CacheManager = cacheManager

	CurrentRepository = repo

	return nil
}

// CriticalEndpointError is returned for SPARQL endpoint responses that must not be retried,
// such as 4xx client errors (including 429 Too Many Requests).
type CriticalEndpointError struct {
	StatusCode int
	message    string
}

func (e *CriticalEndpointError) Error() string {
	return e.message
}

// isCriticalStatusCode reports whether a HTTP status code represents an error that
// retrying cannot resolve. All 4xx client errors are critical; 5xx server errors are retryable.
func isCriticalStatusCode(code int) bool {
	return code >= 400 && code < 500
}

// doQueryCall performs a single HTTP POST to the SPARQL endpoint and returns the response body.
// It returns a CriticalEndpointError for 4xx responses so callers can skip retries.
func (r *Repository) doQueryCall(query, accept string) (*string, error) {
	form := url.Values{}
	form.Set("query", query)
	b := form.Encode()

	req, err := http.NewRequest("POST", r.client.Endpoint, bytes.NewBufferString(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", accept)

	for header, content := range r.client.Headers {
		req.Header.Set(header, content)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received bad(HTTP: " + resp.Status + ") response from SPARQL endpoint:")
		fmt.Println(responseString)
		if isCriticalStatusCode(resp.StatusCode) {
			return nil, &CriticalEndpointError{
				StatusCode: resp.StatusCode,
				message:    "Received critical error response from SPARQL endpoint: " + resp.Status,
			}
		}
		return nil, errors.New("Received bad response from SPARQL endpoint: " + resp.Status)
	}

	return &responseString, nil
}

// QueryCall executes a SPARQL query against the endpoint, retrying transient failures
// when retry is enabled in config. 4xx responses are always treated as critical and
// returned immediately without retrying.
func (r *Repository) QueryCall(query, accept string) (*string, error) {
	retry := r.client.Retry

	maxAttempts := 1
	if retry.Enabled && retry.MaxRetries > 0 {
		maxAttempts = retry.MaxRetries + 1
	}

	initialDelay := retry.InitialDelayMS
	if initialDelay <= 0 {
		initialDelay = 1000
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := time.Duration(initialDelay*(1<<(attempt-1))) * time.Millisecond
			fmt.Printf("Retrying SPARQL request (attempt %d/%d) after %v...\n", attempt+1, maxAttempts, delay)
			time.Sleep(delay)
		}

		result, err := r.doQueryCall(query, accept)
		if err == nil {
			return result, nil
		}

		var criticalErr *CriticalEndpointError
		if errors.As(err, &criticalErr) {
			return nil, err
		}

		lastErr = err
		if retry.Enabled && attempt < maxAttempts-1 {
			fmt.Printf("SPARQL request failed (attempt %d/%d): %v\n", attempt+1, maxAttempts, err)
		}
	}
	return nil, lastErr
}

// Get a query contents by its filename and optionally substitute variables.
func (r *Repository) queryBody(queryLocation string, arguments ...interface{}) (*string, error) {
	query, exists := r.QueryIndex[queryLocation] // QueryIndex includes query/, wanted or not? not?
	if !exists {
		return nil, errors.New("The given query could not be found. " + queryLocation)
	}

	if len(arguments) > 0 {
		for _, argument := range arguments {
			argument := cast.ToString(argument)
			query = strings.Replace(query, "{{.}}", argument, 1)
		}
	}

	if r.verbose {
		if len(arguments) > 0 {
			promt := fmt.Sprintf("Issuing parameterized query %v with arguments: %v.", queryLocation, arguments)
			fmt.Println(promt)
		} else {
			fmt.Println("Issuing query: " + queryLocation)
		}
	}

	return &query, nil
}

// Retrieve the query results from cache or issue a new call and save to cache
func (r *Repository) queryCallCached(queryLocation, query, accept string) (io.Reader, error) {
	file, err := r.CacheManager.GetCache(queryLocation, query)
	if err != nil {
		return nil, err
	}

	if file != nil {
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		return strings.NewReader(string(content)), nil
	}

	queryResult, err := r.QueryCall(query, accept)
	if err != nil {
		return nil, err
	}

	if err := r.CacheManager.SetCache(queryLocation, query, *queryResult); err != nil {
		return nil, err
	}

	var resultReader = strings.NewReader(*queryResult)
	return resultReader, nil
}

func (r *Repository) QuerySelect(queryLocation string, arguments ...interface{}) ([]map[string]rdf.Term, error) {
	accept := "application/sparql-results+json"

	queryBody, err := r.queryBody(queryLocation, arguments...)
	if err != nil {
		return nil, err
	}

	responseContents, err := r.queryCallCached(queryLocation, *queryBody, accept)
	if err != nil {
		return nil, err
	}

	parsedResponse, err := ParseSPARQLJSON(responseContents)
	if err != nil {
		return nil, err
	}
	return parsedResponse, nil
}

func (r *Repository) QueryConstruct(queryLocation string, arguments ...interface{}) (interface{}, error) {
	accept := "application/ld+json"

	queryBody, err := r.queryBody(queryLocation, arguments...)
	if err != nil {
		return nil, err
	}

	responseContents, err := r.queryCallCached(queryLocation, *queryBody, accept)
	if err != nil {
		return nil, err
	}

	var parsedResponse interface{}

	err = json.NewDecoder(responseContents).Decode(&parsedResponse)
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

type Results struct {
	Variables []string
	Results   results
}

type results struct {
	Bindings []map[string]binding
}

type binding struct {
	Type     string
	Value    string
	Lang     string `json:"xml:lang"`
	DataType string
}

var xsdString, _ = rdf.NewIRI("http://www.w3.org/2001/XMLSchema#string")

func ParseSPARQLJSON(r io.Reader) ([]map[string]rdf.Term, error) {
	var results Results
	err := json.NewDecoder(r).Decode(&results)
	if err != nil {
		return nil, err
	}

	var parsedResults []map[string]rdf.Term
	for _, binding := range results.Results.Bindings {
		parsedBinding := make(map[string]rdf.Term)
		for key, value := range binding {
			var term rdf.Term
			var err error
			switch value.Type {
			case "bnode":
				term, err = rdf.NewBlank(value.Value)
			case "uri":
				term, err = rdf.NewIRI(value.Value)
			case "literal":
				// Untyped literals are typed as xsd:string
				if value.Lang != "" {
					term, err = rdf.NewLangLiteral(value.Value, value.Lang)
					break
				}

				if value.DataType != "" {
					iri, _ := rdf.NewIRI(value.DataType)
					term = rdf.NewTypedLiteral(value.Value, iri)
					break
				}
				term = rdf.NewTypedLiteral(value.Value, xsdString)
			case "typed-literal":
				iri, err := rdf.NewIRI(value.DataType)
				term = rdf.NewTypedLiteral(value.Value, iri)
				if err != nil {
					term = nil
					err = nil
				}
			default:
				term = nil
				err = errors.New("Unknown RDF type")
			}

			if err == nil {
				parsedBinding[key] = term
			}
		}
		parsedResults = append(parsedResults, parsedBinding)
	}

	return parsedResults, nil
}
