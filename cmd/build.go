package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/config"
	"github.com/glaciers-in-archives/snowman/internal/rdf"
	"github.com/glaciers-in-archives/snowman/internal/sparql"
	"github.com/glaciers-in-archives/snowman/internal/static"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/glaciers-in-archives/snowman/internal/views"
	"github.com/spf13/cobra"
)

// CLI FLAGS
var sparqlCacheBuildOption string
var resourcesCacheBuildOption string
var staticBuildOption bool
var configFileLocation string
var snowmanDirectoryPath string

func DiscoverLayouts() ([]string, error) {
	var paths []string
	fs.WalkDir(os.DirFS("."), "templates/layouts", func(path string, info fs.DirEntry, err error) error {
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
		printVerbose("Failed to locate query files. Skipping...")
		return index, nil
	}

	err := fs.WalkDir(os.DirFS("."), "queries", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sparqlBytes, err := os.ReadFile(path)
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

			printVerbose("Finished updating static files.")
			return nil
		}

		err := config.LoadConfig(configFileLocation)
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

		sparqlCacheManager, err := cache.NewSparqlCacheManager(sparqlCacheBuildOption, snowmanDirectoryPath)
		if err != nil {
			return utils.ErrorExit("Failed to initiate cache manager.", err)
		}

		err = sparql.NewRepository(*sparqlCacheManager, queries, verbose)
		if err != nil {
			return utils.ErrorExit("Failed to initiate SPARQL client.", err)
		}

		err = cache.NewResourcesCacheManager(resourcesCacheBuildOption, snowmanDirectoryPath)
		if err != nil {
			return utils.ErrorExit("Failed to initiate resources cache manager.", err)
		}

		discoveredViews, err := views.DiscoverViews(layouts)
		if err != nil {
			return utils.ErrorExit("Failed to discover views.", err)
		}

		if err := os.RemoveAll("site"); err != nil {
			return utils.ErrorExit("Failed to remove the existing site directory.", err)
		}

		if _, err := os.Stat("static"); os.IsNotExist(err) {
			printVerbose("Failed to locate static files. Skipping...")
		} else {
			if err := static.CopyIn(); err != nil {
				return utils.ErrorExit("Failed to copy static files.", err)
			}
			printVerbose("Finished copying static files.")
		}

		var renderedPaths = make(map[string]bool)
		for _, view := range discoveredViews {
			results := make([]map[string]rdf.Term, 0)
			if view.ViewConfig.QueryFile != "" {
				printVerbose("Issuing query " + view.ViewConfig.QueryFile)
				results, err = sparql.CurrentRepository.QuerySelect(view.ViewConfig.QueryFile)
				if err != nil {
					return utils.ErrorExit("SPARQL query failed.", err)
				}
			}

			// if the page is rendered based on SPARQL result rows
			if view.MultipageVariableHook != nil {
				for _, row := range results {
					if _, ok := row[*view.MultipageVariableHook]; !ok {
						err := fmt.Errorf(*view.MultipageVariableHook + " not found in SPARQL result row.")
						return utils.ErrorExit("Failed to render multipage view.", err)
					}

					pathSection := row[*view.MultipageVariableHook].String()
					if utils.ValidatePathSection(pathSection); err != nil {
						return utils.ErrorExit("Failed to validate path section.", err)
					}

					outputVariablePattern := regexp.MustCompile(`{{ *` + *view.MultipageVariableHook + ` *}}`)
					outputPath := "site/" + outputVariablePattern.ReplaceAllString(view.ViewConfig.Output, pathSection)

					if renderedPaths[outputPath] {
						fmt.Println("Warning: Writing to " + outputPath + " for the second time.")
					}

					if err := view.RenderPage(outputPath, row); err != nil {
						return utils.ErrorExit("Failed to render page at "+outputPath, err)
					}
					printVerbose("Rendered page at site/" + outputPath)
					renderedPaths[outputPath] = true
				}
			} else {
				if renderedPaths["site/"+view.ViewConfig.Output] {
					fmt.Println("Warning: Writing to " + "site/" + view.ViewConfig.Output + " for the second time.")
				}

				if err := view.RenderPage("site/"+view.ViewConfig.Output, results); err != nil {
					return utils.ErrorExit("Failed to render page at "+"site/"+view.ViewConfig.Output, err)
				}
				printVerbose("Rendered page at site/" + view.ViewConfig.Output)
				renderedPaths["site/"+view.ViewConfig.Output] = true
			}

		}

		if err := sparql.CurrentRepository.CacheManager.Teardown(); err != nil {
			return utils.ErrorExit("Failed write used queries to cache memory.", err)
		}

		fmt.Println("Finished building project.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVar(&sparqlCacheBuildOption, "cache-sparql", "available", "Sets the cache strategy. \"available\" will use cached SPARQL responses when available and fallback to making queries. \"never\" will ignore existing cache and will not update or set new cache.")
	buildCmd.Flags().StringVar(&resourcesCacheBuildOption, "cache-resources", "available", "Sets the cache strategy for resources. \"available\" will use cached resources when available and fallback to downloading. \"never\" will ignore existing cache and will not update or set new cache.")
	buildCmd.Flags().BoolVarP(&staticBuildOption, "static", "s", false, "When set Snowman will only build static files.")
	buildCmd.Flags().StringVarP(&configFileLocation, "config", "f", "snowman.yaml", "Sets the config file to use.")
	buildCmd.Flags().StringVarP(&snowmanDirectoryPath, "snowman-directory", "d", ".snowman", "Sets the snowman directory to use.")

	buildCmd.Flags().StringVarP(&sparqlCacheBuildOption, "cache", "c", "available", "This flag is deprecated. Use --cache-sparql instead.")
	buildCmd.MarkFlagsMutuallyExclusive("cache", "cache-sparql")
}
