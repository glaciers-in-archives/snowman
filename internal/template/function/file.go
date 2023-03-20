package function

import (
	"errors"
	"html/template"
	"io/ioutil"
)

func readFile(filepath string) (string, error) {
	if filepath[0] == '/' || filepath[0] == '.' || filepath[0] == '~' {
		return "", errors.New("File path must be relative to the project root.")
	}

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func GetFileFuncs() template.FuncMap {
	return map[string]interface{}{
		"read_file": readFile,
	}
}
