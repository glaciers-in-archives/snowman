package config

import (
	"net/url"

	"gopkg.in/yaml.v2"
)

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

func (c *SiteConfig) Get() SiteConfig {
	return *c
}
