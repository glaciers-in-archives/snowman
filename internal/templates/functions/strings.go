package functions

import (
	html_template "html/template"
	"strings"
	text_template "text/template"
)

var stringFuncs = map[string]interface{}{
	"split":   strings.Split,
	"replace": strings.Replace,
	"lcase":   strings.ToLower,
	"ucase":   strings.ToUpper,
	"tcase":   strings.Title,
	"join": func(sep string, strs ...string) string {
		return strings.Join(strs, sep)
	},
}

func GetHTMLStringFuncs() html_template.FuncMap {
	return html_template.FuncMap(stringFuncs)
}

func GetTextStringFuncs() text_template.FuncMap {
	return text_template.FuncMap(stringFuncs)
}
