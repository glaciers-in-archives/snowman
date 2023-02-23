package function

import (
	"encoding/json"
	"html/template"
)

var jsonFuncs = map[string]interface{}{
	"to_json":   toJSON,
	"from_json": fromJSON,
}

func toJSON(arg interface{}) (template.HTML, error) {
	b, err := json.Marshal(arg)

	if err != nil {
		return "", err
	}

	return template.HTML(b), nil
}

func fromJSON(jsonString string) (interface{}, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetJSONFuncs() template.FuncMap {
	return template.FuncMap(jsonFuncs)
}
