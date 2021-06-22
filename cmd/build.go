package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/glaciers-in-archives/snowman/internal/views"
	"github.com/knakk/rdf"
	"github.com/spf13/cobra"
)

// CLI FLAGS
var cacheBuildOption string

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

func DiscoverQueries() (map[string]string, error) {
	var index = make(map[string]string)

	err := filepath.Walk("queries", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sparqlBytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			index[strings.Replace(path, "queries/", "", 1)] = string(sparqlBytes)

		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return index, nil
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

		var siteConfig config.SiteConfig
		if err := siteConfig.Parse(data); err != nil {
			return utils.ErrorExit("Failed to parse snowman.yaml.", err)
		}

		templates, err := DiscoverTemplates()
		if err != nil {
			return utils.ErrorExit("Failed to find any template files.", err)
		}

		if len(templates) == 0 {
			return errors.New("Failed to find any template files.")
		}

		queries, err := DiscoverQueries()
		if err != nil {
			return utils.ErrorExit("Failed to find any query files.", err)
		}

		repo, err := sparql.NewRepository(siteConfig.Endpoint, http.DefaultClient, cacheBuildOption, queries)
		if err != nil {
			return utils.ErrorExit("Failed to initiate SPARQL client.", err)
		}

		discoveredViews, err := views.DiscoverViews(templates, *repo, siteConfig)
		if err != nil {
			return utils.ErrorExit("Failed to discover views.", err)
		}

		var siteDir string = "site/"

		if _, err := os.Stat("static"); os.IsNotExist(err) {
			fmt.Println("Failed to locate static files. Skipping...")
		} else {
			if err := utils.CopyDir("static", "site"); err != nil {
				return utils.ErrorExit("Failed to copy static files.", err)
			}
			fmt.Println("Finished copying static files.")
		}

		for _, view := range discoveredViews {
			results := make([]map[string]rdf.Term, 0)
			if view.ViewConfig.QueryFile != "" {
				fmt.Println("Issuing query " + view.ViewConfig.QueryFile)
				results, err = repo.Query(view.ViewConfig.QueryFile)
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
	buildCmd.Flags().StringVarP(&cacheBuildOption, "cache", "c", "available", "Sets the cache strategy. \"available\" will use cached SPARQL responses when available and fallback to making queries. \"never\" will ignore existing cache and will not update or set new cache.")
}
