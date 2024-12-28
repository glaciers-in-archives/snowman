package function

import (
	"errors"
	"os"
)

func ReadFile(filepath string) (string, error) {
	if filepath[0] == '/' || filepath[0] == '.' || filepath[0] == '~' {
		return "", errors.New("File path must be relative to the project root.")
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
