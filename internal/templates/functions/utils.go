package functions

import (
	"html/template"
	html_template "html/template"
	"os"
	text_template "text/template"
	"time"

	"github.com/knakk/rdf"
)

var utilFuncs = map[string]interface{}{
	"now": time.Now,
	"env": os.Getenv,
	"safe_html": func(str string) template.HTML {
		return template.HTML(str)
	},
	"uri": func(value string) (rdf.IRI, error) {
		return rdf.NewIRI(value)
	},
}

func GetHTMLUtilsFuncs() html_template.FuncMap {
	return html_template.FuncMap(utilFuncs)
}

func GetTextUtilsFuncs() text_template.FuncMap {
	return text_template.FuncMap(utilFuncs)
}
