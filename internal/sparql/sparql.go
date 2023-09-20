package sparql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/knakk/rdf"
	"github.com/spf13/cast"
)

type Repository struct {
	client       config.ClientConfig
	httpClient   *http.Client
	verbose      bool
	CacheManager *cache.CacheManager
	QueryIndex   map[string]string
}

var CurrentRepository Repository

func NewRepository(cacheStrategy string, queryIndex map[string]string, verbose bool) error {
	repo := Repository{
		client:     config.CurrentSiteConfig.Client,
		QueryIndex: queryIndex,
		verbose:    verbose,
	}
	repo.httpClient = http.DefaultClient

	cm, err := cache.NewCacheManager(cacheStrategy)
	if err != nil {
		return errors.New("Failed to initiate cache handler. " + " Error: " + err.Error())
	}

	repo.CacheManager = cm

	CurrentRepository = repo

	return nil
}

func (r *Repository) QueryCall(query string) (*string, error) {
	form := url.Values{}
	form.Set("query", query)
	b := form.Encode()

	req, err := http.NewRequest("POST", r.client.Endpoint, bytes.NewBufferString(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", "application/sparql-results+json")

	for header, content := range r.client.Headers {
		req.Header.Set(header, content)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received bad(HTTP: " + resp.Status + ") response from SPARQL endpoint:")
		fmt.Println(responseString)
		return nil, errors.New("Received bad response from SPARQL endpoint")
	}

	return &responseString, nil
}

func (r *Repository) Query(queryLocation string, arguments ...interface{}) ([]map[string]rdf.Term, error) {
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

	file, err := r.CacheManager.GetCache(queryLocation, query)
	if err != nil {
		return nil, err
	}

	var parsedResponse []map[string]rdf.Term
	if file != nil {
		parsedResponse := ParseSPARQLJSON(file)
		if err != nil {
			return nil, err
		}

		file.Close()
		return parsedResponse, nil
	}

	jsonString, err := r.QueryCall(query)
	if err != nil {
		return nil, err
	}

	if err := r.CacheManager.SetCache(queryLocation, query, *jsonString); err != nil {
		return nil, err
	}

	var resultReader = strings.NewReader(*jsonString)
	parsedResponse = ParseSPARQLJSON(resultReader)
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

func ParseSPARQLJSON(r io.Reader) []map[string]rdf.Term {
	var results Results
	err := json.NewDecoder(r).Decode(&results)

	if err != nil {
		return nil
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

	return parsedResults
}
