package sparql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
}

func NewRepository(endpoint string, client *http.Client, cacheStrategy string) (*Repository, error) {
	repo := Repository{
		Endpoint: endpoint,
		Client:   http.DefaultClient,
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

func (r *Repository) Query(queryLocation string, query string) ([]map[string]rdf.Term, error) {
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

func (r *Repository) DynamicQuery(queryLocation string, argument string) ([]map[string]rdf.Term, error) {
	fmt.Println("Issuing dynamic query " + queryLocation + " with argument " + argument)
	queryPath := "queries/" + queryLocation + ".rq"
	if _, err := os.Stat(queryPath); err != nil {
		return nil, err
	}

	sparqlBytes, err := ioutil.ReadFile(queryPath)
	if err != nil {
		return nil, err
	}

	sparqlString := strings.Replace(string(sparqlBytes), "{{.}}", argument, 1)
	parsedResponse, err := r.Query(queryLocation, sparqlString)
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}
