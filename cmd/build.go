package cmd

import (
	"errors"
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
	"github.com/glaciers-in-archives/snowman/internal/version"
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

// renderView issues the query of a view and renders its page(s). If argument
// isn't nil, only the pages of a multipage view where the variable used in the
// output path has that value are rendered. It returns the number of rendered pages.
func renderView(view views.View, argument *string, renderedPaths map[string]bool) (int, error) {
	var err error
	results := make([]map[string]rdf.Term, 0)
	if view.ViewConfig.QueryFile != "" {
		printVerbose("Issuing query " + view.ViewConfig.QueryFile)
		results, err = sparql.CurrentRepository.QuerySelect(view.ViewConfig.QueryFile)
		if err != nil {
			return 0, utils.ErrorExit("SPARQL query failed.", err)
		}
	}

	// if the page is rendered based on all of the SPARQL results
	if view.MultipageVariableHook == nil {
		outputPath := "site/" + view.ViewConfig.Output
		if renderedPaths[outputPath] {
			fmt.Println("Warning: Writing to " + outputPath + " for the second time.")
		}

		if err := view.RenderPage(outputPath, results); err != nil {
			return 0, utils.ErrorExit("Failed to render page at "+outputPath, err)
		}
		printVerbose("Rendered page at " + outputPath)
		renderedPaths[outputPath] = true
		return 1, nil
	}

	// if the page is rendered based on SPARQL result rows
	printVerbose("Rendering multipage view based on variable " + *view.MultipageVariableHook + ", with " + fmt.Sprint(len(results)) + " results.")
	outputVariablePattern := regexp.MustCompile(`{{ *` + *view.MultipageVariableHook + ` *}}`)
	renderedPages := 0
	for _, row := range results {
		if _, ok := row[*view.MultipageVariableHook]; !ok {
			err := fmt.Errorf(*view.MultipageVariableHook + " not found in SPARQL result row.")
			return renderedPages, utils.ErrorExit("Failed to render multipage view.", err)
		}

		pathSection := row[*view.MultipageVariableHook].String()
		if argument != nil && pathSection != *argument {
			continue
		}

		if utils.ValidatePathSection(pathSection); err != nil {
			return renderedPages, utils.ErrorExit("Failed to validate path section.", err)
		}

		outputPath := "site/" + outputVariablePattern.ReplaceAllString(view.ViewConfig.Output, pathSection)

		if renderedPaths[outputPath] {
			fmt.Println("Warning: Writing to " + outputPath + " for the second time.")
		}

		if err := view.RenderPage(outputPath, row); err != nil {
			return renderedPages, utils.ErrorExit("Failed to render page at "+outputPath, err)
		}
		printVerbose("Rendered page at " + outputPath)
		renderedPaths[outputPath] = true
		renderedPages++
	}

	if argument != nil && renderedPages == 0 {
		return 0, fmt.Errorf("Failed to render multipage view. Error: %s did not return any result where %s is %s.", view.ViewConfig.QueryFile, *view.MultipageVariableHook, *argument)
	}

	return renderedPages, nil
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [view] [argument]",
	Short: "Builds a Snowman site in the current directory.",
	Long: `Tries to locate the Snowman configuration, views, queries, etc in the current directory. Then tries to build a Snowman site.

To only build a part of the site, provide the output of a view as it's written in views.yaml. For a view which renders one page per SPARQL result, a second argument limits the build to the page(s) where the variable used in the output path has the given value. A partial build leaves the rest of the already built site in place.`,
	Example: `  snowman build
  snowman build index.html
  snowman build "entities/{{id}}.html"
  snowman build "entities/{{id}}.html" Q42`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		// a partial build only renders the given view, optionally limited to the given argument
		partialBuild := len(args) > 0
		var viewSelector string
		var viewArgument *string
		if partialBuild {
			if staticBuildOption {
				return errors.New("--static can't be combined with a view.")
			}
			if args[0] == "" {
				return errors.New("The view to build can't be empty.")
			}
			viewSelector = args[0]
			if len(args) == 2 {
				viewArgument = &args[1]
			}
		}

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

		// Check version compatibility - block build if incompatible
		if !config.CheckVersionCompatibility() {
			if config.CurrentSiteConfig.SnowmanVersion != "" {
				return fmt.Errorf("Error: Version mismatch. Your Snowman version (%s) does not satisfy the project requirement (%s)",
					version.CurrentVersion.String(),
					config.CurrentSiteConfig.SnowmanVersion)
			}
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

		discoveredViews, err := views.DiscoverViews(layouts, viewSelector)
		if err != nil {
			return utils.ErrorExit("Failed to discover views.", err)
		}

		if viewArgument != nil {
			for _, view := range discoveredViews {
				if view.MultipageVariableHook == nil {
					return fmt.Errorf("The view %s renders a single page and doesn't take an argument.", view.ViewConfig.Output)
				}
			}
		}

		// a partial build leaves everything but the pages it renders in place
		if !partialBuild {
			if err := utils.EmptyDir("site"); err != nil {
				return utils.ErrorExit("Failed to empty the existing site directory.", err)
			}

			if _, err := os.Stat("static"); os.IsNotExist(err) {
				printVerbose("Failed to locate static files. Skipping...")
			} else {
				if err := static.CopyIn(); err != nil {
					return utils.ErrorExit("Failed to copy static files.", err)
				}
				printVerbose("Finished copying static files.")
			}
		}

		var renderedPaths = make(map[string]bool)
		renderedPages := 0
		for _, view := range discoveredViews {
			pages, err := renderView(view, viewArgument, renderedPaths)
			if err != nil {
				return err
			}
			renderedPages += pages
		}

		if err := sparql.CurrentRepository.CacheManager.Teardown(partialBuild); err != nil {
			return utils.ErrorExit("Failed write used queries to cache memory.", err)
		}

		if partialBuild {
			fmt.Println("Finished partial build, rendered " + fmt.Sprint(renderedPages) + " page(s).")
		} else {
			fmt.Println("Finished building project.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVar(&sparqlCacheBuildOption, "cache-sparql", "available", "Sets the cache strategy. \"available\" will use cached SPARQL responses when available and fallback to making queries. \"never\" will ignore existing cache and will not update or set new cache.")
	buildCmd.Flags().StringVar(&resourcesCacheBuildOption, "cache-resources", "available", "Sets the cache strategy for resources. \"available\" will use cached resources when available and fallback to downloading. \"never\" will ignore existing cache and will not update or set new cache.")
	buildCmd.Flags().BoolVarP(&staticBuildOption, "static", "s", false, "Only process files in the \"static\" directory, otherwise leave already built site in place.")
	buildCmd.Flags().StringVarP(&configFileLocation, "config", "f", "snowman.yaml", "Sets the config file to use.")
	buildCmd.Flags().StringVarP(&snowmanDirectoryPath, "snowman-directory", "d", ".snowman", "Sets the snowman directory to use.")

	buildCmd.Flags().StringVarP(&sparqlCacheBuildOption, "cache", "c", "available", "This flag is deprecated. Use --cache-sparql instead.")
	buildCmd.MarkFlagsMutuallyExclusive("cache", "cache-sparql")
}
