package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/glaciers-in-archives/snowman/internal/static"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/glaciers-in-archives/snowman/internal/views"
	"github.com/knakk/rdf"
	"github.com/spf13/cobra"
)

// CLI FLAGS
var cacheBuildOption string
var staticBuildOption bool

func DiscoverLayouts() ([]string, error) {
	var paths []string
	filepath.Walk("templates/layouts", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, nil
}

func DiscoverQueries() (map[string]string, error) {
	var index = make(map[string]string)

	if _, err := os.Stat("queries"); os.IsNotExist(err) {
		fmt.Println("Failed to locate query files. Skipping...")
		return index, nil
	}

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

		if staticBuildOption {
			if err := static.ClearStatic(); err != nil {
				utils.ErrorExit("Failed to clear old static files: ", err)
			}

			if err := static.CopyIn(); err != nil {
				utils.ErrorExit("Failed to copy new static files: ", err)
			}

			fmt.Println("Finished updating static files.")
			return nil
		}

		err := config.LoadConfig()
		if err != nil {
			return err
		}

		layouts, err := DiscoverLayouts()
		if err != nil {
			return utils.ErrorExit("Failed to find any template files.", err)
		}

		queries, err := DiscoverQueries()
		if err != nil {
			return utils.ErrorExit("Failed to index query files.", err)
		}

		err = sparql.NewRepository(cacheBuildOption, queries)
		if err != nil {
			return utils.ErrorExit("Failed to initiate SPARQL client.", err)
		}

		discoveredViews, err := views.DiscoverViews(layouts)
		if err != nil {
			return utils.ErrorExit("Failed to discover views.", err)
		}

		if _, err := os.Stat("site"); err != nil {
			if err := os.RemoveAll("site"); err != nil {
				return utils.ErrorExit("Failed to remove the existing site directory.", err)
			}
		}

		if _, err := os.Stat("static"); os.IsNotExist(err) {
			fmt.Println("Failed to locate static files. Skipping...")
		} else {
			if err := static.CopyIn(); err != nil {
				return utils.ErrorExit("Failed to copy static files.", err)
			}
			fmt.Println("Finished copying static files.")
		}

		var renderedPaths = make(map[string]bool)
		for _, view := range discoveredViews {
			results := make([]map[string]rdf.Term, 0)
			if view.ViewConfig.QueryFile != "" {
				fmt.Println("Issuing query " + view.ViewConfig.QueryFile)
				results, err = sparql.CurrentRepository.Query(view.ViewConfig.QueryFile)
				if err != nil {
					return utils.ErrorExit("SPARQL query failed.", err)
				}
			}

			if view.MultipageVariableHook != nil {
				for _, row := range results {
					outputPath := "site/" + strings.Replace(view.ViewConfig.Output, "{{"+*view.MultipageVariableHook+"}}", row[*view.MultipageVariableHook].String(), 1)

					if renderedPaths[outputPath] {
						fmt.Println("Warning: Writing to " + outputPath + " for the second time.")
					}

					if err := view.RenderPage(outputPath, row); err != nil {
						return utils.ErrorExit("Failed to render page at "+outputPath, err)
					}
					renderedPaths[outputPath] = true
				}
			} else {
				if renderedPaths["site/"+view.ViewConfig.Output] {
					fmt.Println("Warning: Writing to " + "site/" + view.ViewConfig.Output + " for the second time.")
				}

				if err := view.RenderPage("site/"+view.ViewConfig.Output, results); err != nil {
					return utils.ErrorExit("Failed to render page at "+"site/"+view.ViewConfig.Output, err)
				}
				renderedPaths["site/"+view.ViewConfig.Output] = true
			}

		}

		if err := sparql.CurrentRepository.CacheManager.Teardown(); err != nil {
			return utils.ErrorExit("Failed write used queries to cache memory.", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&cacheBuildOption, "cache", "c", "available", "Sets the cache strategy. \"available\" will use cached SPARQL responses when available and fallback to making queries. \"never\" will ignore existing cache and will not update or set new cache.")
	buildCmd.Flags().BoolVarP(&staticBuildOption, "static", "s", false, "When set Snowman will only build static files.")
}
