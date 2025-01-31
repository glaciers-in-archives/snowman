package function

import (
	"fmt"
	"html/template"

	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/rdf"
	"github.com/glaciers-in-archives/snowman/internal/version"
	"github.com/spf13/cast"
)

func SafeHTML(str interface{}) template.HTML {
	return template.HTML(cast.ToString(str))
}

func URI(value string) (rdf.IRI, error) {
	return rdf.NewIRI(value)
}

func Config() config.SiteConfig {
	return config.CurrentSiteConfig
}

func Version() string {
	return version.CurrentVersion.String()
}

func Type(variable interface{}) string {
	return fmt.Sprintf("%T", variable)
}
