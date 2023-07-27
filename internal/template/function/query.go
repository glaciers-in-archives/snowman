package function

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/knakk/rdf"
	"github.com/spf13/cast"
)

var queryFuncs = map[string]interface{}{
	"query": func(queryLocation string, arguments ...interface{}) ([]map[string]rdf.Term, error) {
		query, exists := sparql.CurrentRepository.QueryIndex[queryLocation]
		if !exists {
			return nil, errors.New("The given query could not be found. " + queryLocation)
		}

		switch len(arguments) {
		case 0:
			return sparql.CurrentRepository.Query(queryLocation)
		default:
			var sparqlString = query
			for _, argument := range arguments {
				argument := cast.ToString(argument)
				sparqlString = strings.Replace(sparqlString, "{{.}}", argument, 1)
			}
			promt := fmt.Sprintf("Issuing parameterized query %v with arguments: %v.", queryLocation, arguments)
			fmt.Println(promt)
			return sparql.CurrentRepository.Query(queryLocation, sparqlString)
		}
	},
}

func GetQueryFuncs() template.FuncMap {
	return template.FuncMap(queryFuncs)
}
