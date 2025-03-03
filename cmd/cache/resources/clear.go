package resources_cache_cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

var resourcesCacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the Snowman resources cache.",
	Long:  `This command allows you to clear the cache for any cached non-SPARQL request. The first argument should be the name of the URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := cache.NewResourcesCacheManager("available", snowmanPath)
		if err != nil {
			return utils.ErrorExit("Failed to create cache manager.", err)
		}

		cacheLocation := cache.CurrentResourcesCacheManager.SnowmanDirectoryPath + "/cache/resources/"

		if len(args) == 0 && !unusedOption {
			files, err := os.ReadDir(cacheLocation)
			if err != nil {
				return utils.ErrorExit("Failed to read directory: ", err)
			}

			// TODO: this could become a utility function
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
			unusedCacheItems, err := cache.CurrentResourcesCacheManager.GetUnusedCacheHashes()
			if err != nil {
				return utils.ErrorExit("Failed to get unused cache items.", err)
			}

			for _, item := range unusedCacheItems {
				fmt.Println("Removing: " + item)
				if err := os.RemoveAll(item); err != nil {
					return utils.ErrorExit("Failed to remove the cache file.", err)
				}
			}
		} else if len(args) == 1 {
			itemFilePath := cacheLocation + cache.Hash(args[0])
			directoryPath := filepath.Dir(strings.Split(itemFilePath, "/")[1])

			// remove the file
			if err := os.RemoveAll(itemFilePath); err != nil {
				return utils.ErrorExit("Failed to remove the cache file.", err)
			}

			// if the directory is empty, remove it
			files, err := os.ReadDir(directoryPath)
			if err != nil {
				return utils.ErrorExit("Failed to read directory: ", err)
			}

			if len(files) == 0 {
				if err := os.RemoveAll(directoryPath); err != nil {
					return utils.ErrorExit("Failed to remove the cache directory.", err)
				}
			}
		}

		return nil
	},
}
