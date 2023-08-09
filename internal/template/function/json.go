package function

import (
	"encoding/json"
	"html/template"
)

func ToJSON(arg interface{}) (template.HTML, error) {
	b, err := json.Marshal(arg)

	if err != nil {
		return "", err
	}

	return template.HTML(b), nil
}

func FromJSON(jsonString string) (interface{}, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
