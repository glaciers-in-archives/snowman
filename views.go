package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type View struct {
	ViewConfig   viewConfig
	Sparql       string
	Template     *template.Template
	TemplateName string
}

type viewConfig struct {
	Output       string `yaml:"output"`
	QueryFile    string `yaml:"query"`
	TemplateFile string `yaml:"template"`
}

func (c *viewConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, &c)
}

func DiscoverViews(layoutsTemplate *template.Template) ([]View, error) {
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

				HTMLTemplate, err := layoutsTemplate.ParseFiles(templatePath)
				if err != nil {
					return err
				}

				view := View{
					ViewConfig:   vConfig,
					Sparql:       string(sparqlBytes),
					Template:     HTMLTemplate,
					TemplateName: file,
				}
				views = append(views, view)
			}
		}
		return err
	})
	return views, err
}
