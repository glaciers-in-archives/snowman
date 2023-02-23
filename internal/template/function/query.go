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
		case 1:
			argument := cast.ToString(arguments[0])
			fmt.Println("Issuing parameterized query " + queryLocation + " with argument \"" + argument + "\".")
			sparqlString := strings.Replace(query, "{{.}}", argument, 1)
			return sparql.CurrentRepository.Query(queryLocation, sparqlString)
		}

		return nil, errors.New("Invalid arguments.")
	},
}

func GetQueryFuncs() template.FuncMap {
	return template.FuncMap(queryFuncs)
}
