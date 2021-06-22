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
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
)

type Repository struct {
	Endpoint     string
	Client       *http.Client
	CacheManager *cache.CacheManager
	QueryIndex   map[string]string
}

func NewRepository(endpoint string, client *http.Client, cacheStrategy string, queryIndex map[string]string) (*Repository, error) {
	repo := Repository{
		Endpoint:   endpoint,
		Client:     http.DefaultClient,
		QueryIndex: queryIndex,
	}

	cm, err := cache.NewCacheManager(cacheStrategy)
	if err != nil {
		return nil, errors.New("Failed initiate cache handler. " + " Error: " + err.Error())
	}

	repo.CacheManager = cm

	return &repo, nil
}

func (r *Repository) QueryCall(query string) (*string, error) {
	form := url.Values{}
	form.Set("query", query)
	b := form.Encode()

	req, err := http.NewRequest("POST", r.Endpoint, bytes.NewBufferString(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received bad response from SPARQL endpoint.")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseString := string(bodyBytes)

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
