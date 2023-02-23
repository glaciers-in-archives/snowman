package function

import (
	"encoding/json"
	"html/template"
)

var jsonFuncs = map[string]interface{}{
	"to_json": toJSON,
}

func toJSON(arg interface{}) (template.HTML, error) {
	b, err := json.Marshal(arg)

	if err != nil {
		return "", err
	}

	return template.HTML(b), nil
}

func GetJSONFuncs() template.FuncMap {
	return template.FuncMap(jsonFuncs)
}
