package sparql_cache_cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var sparqlCacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the Snowman SPARQL cache.",
	Long:  `This command allows you to clear the cache for any cached query. The first argument should be the name of the SPARQL query. To clear the cache of a parameterized query provide a second argument with its parameter value.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// we do initialize the cache manager, so that all cache related checks are done
		cm, err := cache.NewSparqlCacheManager("available", snowmanPath) // the cache strategy isn't relevant for this command
		if err != nil {
			return utils.ErrorExit("Failed to create cache manager.", err)
		}
		var cacheLocation = cm.SnowmanDirectoryPath + "/cache/sparql/"

		// if we have no arguments and the unused flag is not set, we remove all cache items
		if len(args) == 0 && !unusedOption {
			files, err := os.ReadDir(cacheLocation)
			if err != nil {
				return utils.ErrorExit("Failed to read directory: ", err)
			}

			for _, file := range files {
				if file.IsDir() {
					dirPath := cacheLocation + file.Name()
					err := os.RemoveAll(dirPath)
					if err != nil {
						return utils.ErrorExit("Failed to remove directory: ", err)
					}
				}
			}

		} else if len(args) == 0 && unusedOption {
			unusedCacheItems, err := cm.GetUnusedCacheHashes()
			if err != nil {
				return utils.ErrorExit("Failed to get unused cache items.", err)
			}

			for _, item := range unusedCacheItems {
				fmt.Println("Removing: " + item)
				if err := os.RemoveAll(item); err != nil {
					return utils.ErrorExit("Failed to remove the cache file.", err)
				}
			}
			// if we have one argument, remove all cache items for the query
		} else if len(args) == 1 {
			dirPath := cacheLocation + cache.Hash(args[0])

			files, err := os.ReadDir(dirPath)
			if err != nil {
				return utils.ErrorExit("Failed to read directory: ", err)
			}

			if len(files) > 1 {
				for _, file := range files {
					filePath := dirPath + "/" + file.Name()
					err := os.Remove(filePath)
					if err != nil {
						return utils.ErrorExit("Failed to remove file: ", err)
					}
				}
			} else {
				err := os.RemoveAll(dirPath)
				if err != nil {
					return utils.ErrorExit("Failed to remove directory: ", err)
				}
			}

			// if we have two or more arguments, we show the cache item for the query with the parameter
		} else if len(args) >= 2 {

			sparqlBytes, err := os.ReadFile("queries/" + args[0])
			if err != nil {
				return utils.ErrorExit("Failed to find the query file.", err)
			}

			// for the second and all following arguments
			queryParameters := args[1:]
			query := string(sparqlBytes)
			for _, parameter := range queryParameters {
				argument := cast.ToString(parameter)
				query = strings.Replace(query, "{{.}}", argument, 1)
			}

			filePath := cacheLocation + cache.Hash(args[0]) + "/" + cache.Hash(query) + ".json"

			err = os.Remove(filePath)
			if err != nil {
				return utils.ErrorExit("Failed to remove file: ", err)
			}
		}

		return nil
	},
}
