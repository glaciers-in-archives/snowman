package function

import (
	"encoding/json"
	"html/template"
	html_template "html/template"
	text_template "text/template"
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

func GetHTMLJSONFuncs() html_template.FuncMap {
	return html_template.FuncMap(jsonFuncs)
}

func GetTextJSONFuncs() text_template.FuncMap {
	return text_template.FuncMap(jsonFuncs)
}
