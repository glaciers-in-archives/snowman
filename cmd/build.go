package cmd

import (
	"errors"
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
	"github.com/knakk/rdf"
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

func DiscoverTemplates() ([]string, error) {
	var paths []string
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

		templates, err := DiscoverTemplates()
		if err != nil {
			return utils.ErrorExit("Failed to find any template files.", err)
		}

		if len(templates) == 0 {
			return errors.New("Failed to find any template files.")
		}

		repo, err := sparql.NewRepository(config.Endpoint, http.DefaultClient, cacheBuildOption)
		if err != nil {
			return utils.ErrorExit("Failed to initiate SPARQL client.", err)
		}

		discoveredViews, err := views.DiscoverViews(templates, *repo)
		if err != nil {
			return utils.ErrorExit("Failed to discover views.", err)
		}

		for _, view := range discoveredViews {
			results := make([]map[string]rdf.Term, 0)
			if view.Sparql != "" {
				fmt.Println("Issuing query " + view.ViewConfig.QueryFile)
				results, err = repo.Query(view.ViewConfig.QueryFile, view.Sparql)
				if err != nil {
					return utils.ErrorExit("SPARQL query failed.", err)
				}
			}

			if view.MultipageVariableHook != nil {
				for _, row := range results {
					outputPath := siteDir + strings.Replace(view.ViewConfig.Output, "{{"+*view.MultipageVariableHook+"}}", row[*view.MultipageVariableHook].String(), 1)
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
