package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/knakk/sparql"

	"gopkg.in/yaml.v2"
)

type siteConfig struct {
	Endpoint string `yaml:"sparql_endpoint"`
}

func (c *siteConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func (c siteConfig) IsValid() error {
	_, err := url.ParseRequestURI(c.Endpoint) // #TODO why is https://example valid?
	if err != nil {
		return err
	}
	return nil
}

func ErrorExit(message string, err error) {
	fmt.Println(message, " Error:", err)
	os.Exit(1)
}

func main() {
	fmt.Println("Welcome to Snowman - a static site generator for SPARQL backends.")

	var siteDir string = "site/"
	err := os.Mkdir("site", 0755)
	if err != nil {
		ErrorExit("Failed to create site directory.", err)
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		ErrorExit("Failed to get the current working directory.", err)
	}

	if _, err := os.Stat(currentDirectory + "/snowman.yaml"); err != nil {
		ErrorExit("Unable to locate snowman.yaml in the current working directory.", err)
	}

	data, err := ioutil.ReadFile(currentDirectory + "/snowman.yaml")
	if err != nil {
		ErrorExit("Failed to read snowman.yaml.", err)
	}

	var config siteConfig
	if err := config.Parse(data); err != nil {
		ErrorExit("Failed to parse snowman.yaml.", err)
	}

	if err := config.IsValid(); err != nil {
		ErrorExit("Failed to validate snowman.yaml.", err)
	}

	views, err := DiscoverViews()
	if err != nil {
		ErrorExit("Failed to discover views.", err)
	}

	for _, view := range views {
		repo, err := sparql.NewRepo(config.Endpoint)
		if err != nil {
			ErrorExit("Failed to connect to SPARQL endpoint.", err)
		}
		res, err := repo.Query(view.Sparql)
		if err != nil {
			ErrorExit("SPARQL query failed.", err)
		}

		f, err := os.Create(siteDir + view.ViewConfig.Output)
		if err != nil {
			ErrorExit("Failed to create "+siteDir+view.ViewConfig.Output+" file.", err)
		}
		results := res.Results.Bindings
		view.Template.Execute(os.Stdout, results)
		if err := view.Template.Execute(f, results); err != nil {
			ErrorExit("Failed to render "+siteDir+view.ViewConfig.Output+".", err)
		}
		f.Close()
	}

}
