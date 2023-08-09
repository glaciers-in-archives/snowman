package function

import (
	"errors"
	"io/ioutil"
)

func ReadFile(filepath string) (string, error) {
	if filepath[0] == '/' || filepath[0] == '.' || filepath[0] == '~' {
		return "", errors.New("File path must be relative to the project root.")
	}

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
