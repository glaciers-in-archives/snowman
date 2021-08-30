package function

import (
	"errors"
	"fmt"
	html_template "html/template"
	"strings"
	text_template "text/template"

	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/knakk/rdf"
)

var queryFuncs = map[string]interface{}{
	"query": func(queryLocation string, arguments ...string) ([]map[string]rdf.Term, error) {
		if !strings.HasSuffix(queryLocation, ".rq") {
			queryLocation += ".rq"
		}

		query, exists := sparql.CurrentRepository.QueryIndex[queryLocation]
		if !exists {
			return nil, errors.New("The given query could not be found. " + queryLocation)
		}

		switch len(arguments) {
		case 0:
			return sparql.CurrentRepository.Query(queryLocation)
		case 1:
			fmt.Println("Issuing parameterized query " + queryLocation + " with argument \"" + arguments[0] + "\".")
			sparqlString := strings.Replace(query, "{{.}}", arguments[0], 1)
			return sparql.CurrentRepository.Query(queryLocation, sparqlString)
		}

		return nil, errors.New("Invalid arguments.")
	},
}

func GetHTMLQueryFuncs() html_template.FuncMap {
	return html_template.FuncMap(queryFuncs)
}

func GetTextQueryFuncs() text_template.FuncMap {
	return text_template.FuncMap(queryFuncs)
}
