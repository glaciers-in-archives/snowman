package resources_cache_cmd

import (
	"fmt"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

var resourcesCacheInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect the Snowman resources cache.",
	Long:  `This command allows you to inspect the cache for any cached non-SPARQL request. The first argument should be the name of the SPARQL query. To inspect the cache of a parameterized query provide a second argument with its parameter value and only argument should be the URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := cache.NewResourcesCacheManager("available", snowmanPath)
		if err != nil {
			return utils.ErrorExit("Failed to create cache manager.", err)
		}

		cacheLocation := cache.CurrentResourcesCacheManager.SnowmanDirectoryPath + "/cache/resources/"

		if len(args) == 0 && !unusedOption {
			totFiles, err := utils.CountFilesRecursive(cacheLocation)
			if err != nil {
				return utils.ErrorExit("Failed to retrive cache info.", err)
			}

			fmt.Println("There are " + fmt.Sprint(totFiles) + " cache items.")
		} else if len(args) == 0 && unusedOption {
			unusedCacheItems, err := cache.CurrentResourcesCacheManager.GetUnusedCacheHashes()
			if err != nil {
				return utils.ErrorExit("Failed to get unused cache items.", err)
			}

			fmt.Println("Found " + fmt.Sprint(len(unusedCacheItems)) + " unused cache items.")
		} else if len(args) == 1 {
			filePath := cacheLocation + cache.Hash(args[0])

			err := utils.PrintFileContents(filePath)
			if err != nil {
				return utils.ErrorExit("Failed to print file contents.", err)
			}
		}

		return nil
	},
}
