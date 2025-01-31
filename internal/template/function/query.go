package function

import (
	"github.com/glaciers-in-archives/snowman/internal/rdf"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
)

func Query(queryLocation string, arguments ...interface{}) ([]map[string]rdf.Term, error) {
	if len(arguments) == 0 {
		return sparql.CurrentRepository.Query(queryLocation)
	}

	return sparql.CurrentRepository.Query(queryLocation, arguments...)
}
