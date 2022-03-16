package function

import (
	html_template "html/template"
	"strings"
	text_template "text/template"

	"github.com/spf13/cast"
)

var stringFuncs = map[string]interface{}{
	"split":      split,
	"replace":    replace,
	"lcase":      lcase,
	"ucase":      ucase,
	"tcase":      tcase,
	"join":       join,
	"has_prefix": has_prefix,
	"has_suffix": has_suffix,
}

func split(str interface{}, sep interface{}) []string {
	return strings.Split(cast.ToString(str), cast.ToString(sep))
}

func replace(str interface{}, old interface{}, new interface{}, count interface{}) string {
	return strings.Replace(cast.ToString(str), cast.ToString(old), cast.ToString(new), cast.ToInt(count))
}

func lcase(str interface{}) string {
	return strings.ToLower(cast.ToString(str))
}

func ucase(str interface{}) string {
	return strings.ToUpper(cast.ToString(str))
}

func tcase(str interface{}) string {
	return strings.Title(cast.ToString(str))
}

func has_prefix(str interface{}, prefix interface{}) bool {
	return strings.HasPrefix(cast.ToString(str), cast.ToString(prefix))
}

func has_suffix(str interface{}, suffix interface{}) bool {
	return strings.HasSuffix(cast.ToString(str), cast.ToString(suffix))
}

func join(sep interface{}, strs ...interface{}) string {
	return strings.Join(cast.ToStringSlice(strs), cast.ToString(sep))
}

func GetHTMLStringFuncs() html_template.FuncMap {
	return html_template.FuncMap(stringFuncs)
}

func GetTextStringFuncs() text_template.FuncMap {
	return text_template.FuncMap(stringFuncs)
}
