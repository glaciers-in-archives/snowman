package config

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/glaciers-in-archives/snowman/internal/utils"
	"gopkg.in/yaml.v2"
)

var CurrentSiteConfig SiteConfig

type ClientConfig struct {
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"http_headers"`
}

type SiteConfig struct {
	Client   ClientConfig `yaml:"sparql_client"`
	Metadata map[string]interface{}
}

func (c *SiteConfig) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	_, err := url.ParseRequestURI(c.Client.Endpoint) // #TODO why is https://example valid?
	if err != nil {
		return err
	}
	return nil
}

func LoadConfig(fileLocation string) error {
	if _, err := os.Stat(fileLocation); err != nil {
		if fileLocation == "snowman.yaml" {
			return utils.ErrorExit("Unable to locate snowman.yaml in the current working directory.", err)
		} else {
			return utils.ErrorExit("Unable to locate Snowman configuration file at "+fileLocation+".", err)
		}
	}

	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return utils.ErrorExit("Failed to read "+fileLocation+".", err)
	}

	if err := CurrentSiteConfig.Parse(data); err != nil {
		return utils.ErrorExit("Failed to parse "+fileLocation+".", err)
	}

	return nil
}
