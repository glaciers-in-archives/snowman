package function

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
)

var remoteFuncs = map[string]interface{}{
	"get_remote": func(uri string) (*string, error) {
		_, err := url.Parse(uri)
		if err != nil {
			return nil, errors.New("Invalid argument given to get_remote template function.")
		}

		req, err := http.NewRequest("GET", uri, bytes.NewBuffer(nil))
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
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
			fmt.Println("Received bad(HTTP: " + resp.Status + ") response from " + uri + ".")
			fmt.Println(responseString)
			return nil, errors.New("Received bad response from remote resource.")
		}

		return &responseString, nil
	},
}

func GetRemoteFuncs() template.FuncMap {
	return template.FuncMap(remoteFuncs)
}
