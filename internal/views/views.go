package views

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"gopkg.in/yaml.v2"
)

type View struct {
	ViewConfig            viewConfig
	Sparql                string
	Template              *template.Template
	TemplateName          string
	MultipageVariableHook *string
}

func (v *View) RenderPage(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	if err := v.Template.ExecuteTemplate(f, v.TemplateName, data); err != nil {
		return err
	}
	f.Close()

	fmt.Println("Rendered page at " + path)
	return nil
}

type viewConfig struct {
	Output       string `yaml:"output"`
	QueryFile    string `yaml:"query"`
	TemplateFile string `yaml:"template"`
}

func (c *viewConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, &c)
}

func DiscoverViews(includes []string, repo sparql.Repository) ([]View, error) {
	var views []View

	err := filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			if filepath.Ext(path) == ".yaml" {
				viewPath := path
				fmt.Println("Discovered view: ", viewPath)

				data, err := ioutil.ReadFile(viewPath)
				if err != nil {
					return errors.New("Failed to read " + viewPath)
				}

				var vConfig viewConfig
				if err := vConfig.Parse(data); err != nil {
					return errors.New("Failed to parse" + viewPath)
				}

				var multipageVariableHook *string
				re := regexp.MustCompile(`{{([\w\d_]+)}}`)
				if re.Match([]byte(vConfig.Output)) {
					multipageVariableHook = &re.FindAllStringSubmatch(vConfig.Output, 1)[0][1]
				}

				queryPath := "queries/" + vConfig.QueryFile
				if _, err := os.Stat(queryPath); err != nil {
					return errors.New("Unable to find the SPARQL file for the " + viewPath + " view.")
				}

				sparqlBytes, err := ioutil.ReadFile(queryPath)
				if err != nil {
					return err
				}

				templatePath := "templates/" + vConfig.TemplateFile
				if _, err := os.Stat(templatePath); err != nil {
					return errors.New("Unable to find the template file for the " + viewPath + " view.")
				}

				_, file := filepath.Split(templatePath)

				funcMap := template.FuncMap{
					"now":     time.Now,
					"split":   strings.Split,
					"replace": strings.Replace,
					"lcase":   strings.ToLower,
					"ucase":   strings.ToUpper,
					"tcase":   strings.ToTitle,

					"query": repo.DynamicQuery,
				}

				allTemplatePaths := includes
				allTemplatePaths = append(allTemplatePaths, templatePath)
				HTMLTemplate, err := template.New("").Funcs(funcMap).ParseFiles(allTemplatePaths...)
				if err != nil {
					return err
				}

				view := View{
					ViewConfig:            vConfig,
					Sparql:                string(sparqlBytes),
					Template:              HTMLTemplate,
					TemplateName:          file,
					MultipageVariableHook: multipageVariableHook,
				}
				views = append(views, view)
			}
		}
		return err
	})
	return views, err
}
