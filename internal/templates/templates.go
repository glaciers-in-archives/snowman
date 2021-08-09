package templates

import (
	html_template "html/template"
	"os"
	"path/filepath"
	"strings"
	text_template "text/template"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
)

type TemplateCollection struct {
	Repository          sparql.Repository
	ParsedTextTemplates *text_template.Template
	ParsedHTMLTemplates *html_template.Template
	TemplatePaths       []string
}

func DiscoverAndParseTemplates(repo sparql.Repository, siteConfig config.SiteConfig) (*TemplateCollection, error) {
	var paths []string

	var functionMap = map[string]interface{}{
		"now":     time.Now,
		"split":   strings.Split,
		"replace": strings.Replace,
		"lcase":   strings.ToLower,
		"ucase":   strings.ToUpper,
		"tcase":   strings.ToTitle,

		"query":    repo.InlineQuery,
		"config":   siteConfig.Get,
		"metadata": siteConfig.GetMetadata,
	}

	var textTemplateFuncMap = text_template.FuncMap(functionMap)
	var htmlTemplateFuncMap = html_template.FuncMap(functionMap)

	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var parsedTextTemplate *text_template.Template
	parsedTextTemplate, err = text_template.New("").Funcs(textTemplateFuncMap).ParseFiles(paths...)
	if err != nil {
		return nil, err
	}

	var parsedHTMLTemplate *html_template.Template
	parsedHTMLTemplate, err = html_template.New("").Funcs(htmlTemplateFuncMap).ParseFiles(paths...)
	if err != nil {
		return nil, err
	}

	return &TemplateCollection{
		Repository:          repo,
		ParsedTextTemplates: parsedTextTemplate,
		ParsedHTMLTemplates: parsedHTMLTemplate,
		TemplatePaths:       paths,
	}, nil
}
