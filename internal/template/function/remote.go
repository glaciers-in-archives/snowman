package function

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
)

func GetRemoteWithConfig(uri interface{}, config map[interface{}]interface{}) (*string, error) {
	preparedUri := cast.ToString(uri)

	_, err := url.Parse(preparedUri)
	if err != nil {
		return nil, errors.New("Invalid argument given to get_remote template function.")
	}

	req, err := http.NewRequest("GET", preparedUri, bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}

	if config != nil {
		if config["headers"] != nil {
			headers := config["headers"].(map[interface{}]interface{})
			for key, value := range headers {
				req.Header.Set(key.(string), value.(string))
			}
		}
	}

	resp, err := http.DefaultClient.Do(req)
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
		fmt.Println("Received bad(HTTP: " + resp.Status + ") response from " + preparedUri + ".")
		fmt.Println(responseString)
		return nil, errors.New("Received bad response from remote resource.")
	}

	return &responseString, nil
}

func GetRemote(uri interface{}) (*string, error) {
	return GetRemoteWithConfig(uri, nil)
}
