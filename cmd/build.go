package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/glaciers-in-archives/snowman/internal/views"
	knakk_sparql "github.com/knakk/sparql"
	"github.com/spf13/cobra"
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

func DiscoverIncludes() ([]string, error) {
	var paths []string
	err := filepath.Walk("templates/includes", func(path string, info os.FileInfo, err error) error {
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
	return paths, nil
}

func CopyStatic() error {
	// we know from prevous checks that the static folder must exist
	err := filepath.Walk("static", func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			newPath := strings.Replace(path, "static/", "site/", 1)
			if err := os.MkdirAll(filepath.Dir(newPath), 0770); err != nil {
				return err
			}

			original, err := os.Open(path)
			if err != nil {
				return err
			}
			defer original.Close()

			new, err := os.Create(newPath)
			if err != nil {
				return err
			}
			defer new.Close()

			_, err = io.Copy(new, original)
			if err != nil {
				return err
			}
			fmt.Println("Copied static file to: " + newPath)
		}
		return err
	})
	return err
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a Snowman site in the current directory.",
	Long:  `Tries to locate the Snowman configuration, views, queries, etc in the current directory. Then tries to build a Snowman site.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if _, err := os.Stat("snowman.yaml"); err != nil {
			return utils.ErrorExit("Unable to locate snowman.yaml in the current working directory.", err)
		}

		data, err := ioutil.ReadFile("snowman.yaml")
		if err != nil {
			return utils.ErrorExit("Failed to read snowman.yaml.", err)
		}

		var config siteConfig
		if err := config.Parse(data); err != nil {
			return utils.ErrorExit("Failed to parse snowman.yaml.", err)
		}

		if err := config.IsValid(); err != nil {
			return utils.ErrorExit("Failed to validate snowman.yaml.", err)
		}

		var siteDir string = "site/"
		err = os.Mkdir("site", 0755)
		if err != nil {
			return utils.ErrorExit("Failed to create site directory.", err)
		}

		if _, err := os.Stat("static"); os.IsNotExist(err) {
			fmt.Println("Failed to locate static files. Skipping...")
		} else {
			if err := CopyStatic(); err != nil {
				return utils.ErrorExit("Failed to copy static files.", err)
			}
		}

		layouts, err := DiscoverIncludes()
		if err != nil {
			fmt.Println("No includes discovered, skipping.")
			layouts = nil
		}

		discoveredViews, err := views.DiscoverViews(layouts)
		if err != nil {
			return utils.ErrorExit("Failed to discover views.", err)
		}

		var executedQueries = make(map[string]bool, 100)
		for _, view := range discoveredViews {
			repo := sparql.Repository{Endpoint: config.Endpoint, Client: http.DefaultClient}

			if cached == false || executedQueries[view.QueryHash] {
				err := repo.QueryToFile(view.Sparql, ".snowman/cache/"+view.QueryHash+".json")
				if err != nil {
					return utils.ErrorExit("SPARQL query failed.", err)
				}
				executedQueries[view.QueryHash] = true
			}

			reader, err := os.Open(".snowman/cache/" + view.QueryHash + ".json")
			if err != nil {
				return utils.ErrorExit("Failed to read query result from the filesystem.", err)
			}

			parsed_response, err := knakk_sparql.ParseJSON(reader)
			if err != nil {
				return utils.ErrorExit("Failed to parse SPARQL JSON returned by query.", err)
			}

			results := parsed_response.Results.Bindings

			if view.MultipageVariableHook != nil {
				for _, row := range results {
					outputPath := siteDir + strings.Replace(view.ViewConfig.Output, "{{"+*view.MultipageVariableHook+"}}", row[*view.MultipageVariableHook].Value, 1)
					if err := view.RenderPage(outputPath, row); err != nil {
						return utils.ErrorExit("Failed to render page at "+outputPath, err)
					}
				}
			} else {
				if err := view.RenderPage(siteDir+view.ViewConfig.Output, results); err != nil {
					return utils.ErrorExit("Failed to render page at "+siteDir+view.ViewConfig.Output, err)
				}
			}

		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
