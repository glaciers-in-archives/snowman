package function

import (
	"github.com/glaciers-in-archives/snowman/internal/rdf"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
)

func Query(queryLocation string, arguments ...interface{}) ([]map[string]rdf.Term, error) {
	if len(arguments) == 0 {
		return sparql.CurrentRepository.QuerySelect(queryLocation)
	}

	return sparql.CurrentRepository.QuerySelect(queryLocation, arguments...)
}

func QueryConstruct(queryLocation string, arguments ...interface{}) (interface{}, error) {
	if len(arguments) == 0 {
		return sparql.CurrentRepository.QueryConstruct(queryLocation)
	}

	return sparql.CurrentRepository.QueryConstruct(queryLocation, arguments...)
}
