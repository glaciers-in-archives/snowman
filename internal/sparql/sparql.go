package sparql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
)

type Repository struct {
	client       config.ClientConfig
	httpClient   *http.Client
	CacheManager *cache.CacheManager
	QueryIndex   map[string]string
}

var CurrentRepository Repository

func NewRepository(cacheStrategy string, queryIndex map[string]string) error {
	repo := Repository{
		client:     config.CurrentSiteConfig.Client,
		QueryIndex: queryIndex,
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

func (r *Repository) Query(queryLocation string, queryOverride ...string) ([]map[string]rdf.Term, error) {

	query, exists := r.QueryIndex[queryLocation] // QueryIndex includes query/, wanted or not? not?
	if !exists {
		return nil, errors.New("The given query could not be found. " + queryLocation)
	}

	if len(queryOverride) > 0 {
		query = queryOverride[0]
	}

	file, err := r.CacheManager.GetCache(queryLocation, query)
	if err != nil {
		return nil, err
	}

	if file != nil {
		parsedResponse, err := sparql.ParseJSON(file)
		if err != nil {
			return nil, err
		}

		return parsedResponse.Solutions(), nil
	}

	jsonString, err := r.QueryCall(query)
	if err != nil {
		return nil, err
	}

	if err := r.CacheManager.SetCache(queryLocation, query, *jsonString); err != nil {
		return nil, err
	}

	var parsedResponse sparql.Results
	err = json.Unmarshal([]byte(*jsonString), &parsedResponse)
	if err != nil {
		return nil, err
	}

	return parsedResponse.Solutions(), nil
}

func (r *Repository) InlineQuery(queryLocation string, arguments ...string) ([]map[string]rdf.Term, error) {
	if !strings.HasSuffix(queryLocation, ".rq") {
		queryLocation += ".rq"
	}

	query, exists := r.QueryIndex[queryLocation]
	if !exists {
		return nil, errors.New("The given query could not be found. " + queryLocation)
	}

	switch len(arguments) {
	case 0:
		return r.Query(queryLocation)
	case 1:
		fmt.Println("Issuing parameterized query " + queryLocation + " with argument \"" + arguments[0] + "\".")
		sparqlString := strings.Replace(query, "{{.}}", arguments[0], 1)
		return r.Query(queryLocation, sparqlString)
	}

	return nil, errors.New("Invalid arguments.")
}
