package function

import (
	"html/template"

	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/knakk/rdf"
)

var queryFuncs = map[string]interface{}{
	"query": func(queryLocation string, arguments ...interface{}) ([]map[string]rdf.Term, error) {
		if len(arguments) == 0 {
			return sparql.CurrentRepository.Query(queryLocation)
		}

		return sparql.CurrentRepository.Query(queryLocation, arguments...)
	},
}

func GetQueryFuncs() template.FuncMap {
	return template.FuncMap(queryFuncs)
}
