package function

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/version"
	"github.com/knakk/rdf"
	"github.com/spf13/cast"
)

var utilFuncs = map[string]interface{}{
	"now": time.Now,
	"env": os.Getenv,
	"safe_html": func(str interface{}) template.HTML {
		return template.HTML(cast.ToString(str))
	},
	"uri": func(value string) (rdf.IRI, error) {
		return rdf.NewIRI(value)
	},
	"config": func() config.SiteConfig {
		return config.CurrentSiteConfig
	},
	"version": func() string {
		return version.CurrentVersion.String()
	},
	"type": func(variable interface{}) string {
		return fmt.Sprintf("%T", variable)
	},
}

func GetUtilsFuncs() template.FuncMap {
	return template.FuncMap(utilFuncs)
}
