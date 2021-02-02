package sparql

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type Repository struct {
	Endpoint string
	Client   *http.Client
}

func (r *Repository) Query(query string) (*string, error) {

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

func (r *Repository) QueryToFile(query string, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		return err
	}

	jsonBody, err := r.Query(query)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(*jsonBody)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Sync()

	return err
}
