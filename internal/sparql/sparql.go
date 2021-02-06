package sparql

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/knakk/sparql"
)

var CacheLocation string = ".snowman/cache/"

type Repository struct {
	Endpoint     string
	Client       *http.Client
	CacheDefault bool
	CacheHashes  map[string]bool
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

	var responseString string
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received bad response from SPARQL endpoint.")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseString = string(bodyBytes)

	return &responseString, nil
}

func (r *Repository) Query(query string) (*sparql.Results, error) {
	hash := sha256.Sum256([]byte(query))
	hashString := hex.EncodeToString(hash[:])
	queryCacheLocation := CacheLocation + hashString + ".json"

	if !r.CacheHashes[hashString] || !r.CacheDefault {
		jsonBody, err := r.QueryCall(query)
		if err != nil {
			return nil, err
		}

		if err := os.MkdirAll(filepath.Dir(queryCacheLocation), 0770); err != nil {
			return nil, err
		}

		f, err := os.Create(queryCacheLocation)
		if err != nil {
			return nil, err
		}
		_, err = f.WriteString(*jsonBody)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		f.Sync()
	}

	reader, err := os.Open(queryCacheLocation)
	if err != nil {
		return nil, err
	}

	parsedResponse, err := sparql.ParseJSON(reader)
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}
