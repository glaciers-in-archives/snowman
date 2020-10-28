package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type View struct {
	ViewName     string
	ViewConfig   viewConfig
	Sparql       string
	Template     *template.Template
	TemplateFile string
}

type viewConfig struct {
	Output string
}

func (c *viewConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, &c)
}

func DiscoverViews(layoutsTemplate *template.Template) ([]View, error) {
	var views []View
	err := filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			if filepath.Ext(path) == ".yaml" {
				viewName := strings.Replace(path, ".yaml", "", -1)
				fmt.Println("Discovered view: ", viewName)

				if _, err := os.Stat(viewName + ".rq"); err != nil {
					return errors.New("Unable to find the SPARQL file for the " + viewName + " view.")
				}

				sparqlBytes, err := ioutil.ReadFile(viewName + ".rq")
				if err != nil {
					return err
				}

				data, err := ioutil.ReadFile(viewName + ".yaml")
				if err != nil {
					return errors.New("Failed to read snowman.yaml.")
				}

				var vConfig viewConfig
				if err := vConfig.Parse(data); err != nil {
					return errors.New("Failed to parse" + viewName + ".yaml.")
				}

				if _, err := os.Stat(viewName + ".html"); err != nil {
					return errors.New("Unable to find the HTML file for the " + viewName + " view.") // todo work with other formats than html, (check for extra var in config?)
				}

				_, file := filepath.Split(viewName + ".html")

				HTMLTemplate, err := layoutsTemplate.ParseFiles(viewName + ".html")
				if err != nil {
					return err
				}

				view := View{
					ViewName:     viewName,
					ViewConfig:   vConfig,
					Sparql:       string(sparqlBytes),
					Template:     HTMLTemplate,
					TemplateFile: file,
				}
				views = append(views, view)
			}
		}
		return err
	})
	return views, err
}
