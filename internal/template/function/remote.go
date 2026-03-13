package function

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/spf13/cast"
)

var remoteClient = &http.Client{Timeout: 30 * time.Second}

func GetRemoteWithConfig(uri interface{}, config map[interface{}]interface{}) (*string, error) {
	preparedUri := cast.ToString(uri)

	parsed, err := url.ParseRequestURI(preparedUri)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("Invalid argument given to get_remote template function.")
	}

	file, err := cache.CurrentResourcesCacheManager.GetCache(preparedUri)
	if err != nil {
		return nil, err
	}

	if file != nil {
		defer file.Close()
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		fileString := string(fileBytes)
		return &fileString, nil
	}

	req, err := http.NewRequest("GET", preparedUri, bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}

	if config != nil {
		if config["headers"] != nil {
			headers, ok := config["headers"].(map[interface{}]interface{})
			if !ok {
				return nil, errors.New("get_remote: headers must be a map")
			}
			for key, value := range headers {
				k, ok1 := key.(string)
				v, ok2 := value.(string)
				if !ok1 || !ok2 {
					return nil, errors.New("get_remote: header keys and values must be strings")
				}
				req.Header.Set(k, v)
			}
		}
	}

	resp, err := remoteClient.Do(req)
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

	err = cache.CurrentResourcesCacheManager.SetCache(preparedUri, responseString)
	if err != nil {
		return nil, err
	}

	return &responseString, nil
}

func GetRemote(uri interface{}) (*string, error) {
	return GetRemoteWithConfig(uri, nil)
}
