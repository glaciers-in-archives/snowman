package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var invalidateCacheOption bool
var unusedOption bool
var snowmanPath string

func printFileContents(path string) error {
	fmt.Println(path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() error {
		if err = file.Close(); err != nil {
			return err
		}
		return nil
	}()

	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	fmt.Print(string(b))
	return nil
}

// cahceCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Show the contents of cached queries",
	Long:  `This command allows you to inspect the cache for any cached query. The first argument should be the name of the SPARQL query. To inspect the cache of a parameterized query provide a second argument with its parameter value.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedCacheItems []string

		// we do initialize the cache manager, so that all cache related checks are done
		cm, err := cache.NewCacheManager("available", snowmanPath) // the cache strategy isn't relevant for this command
		if err != nil {
			return utils.ErrorExit("Failed to create cache manager.", err)
		}
		var cacheLocation = cm.SnowmanDirectoryPath + "/cache/"

		// if we have no arguments and the unused flag is not set, we just count the cache items
		if len(args) == 0 && !unusedOption {
			totFiles, err := utils.CountFilesRecursive(cacheLocation)
			if err != nil {
				return utils.ErrorExit("Failed to retrive cache info.", err)
			}

			fmt.Println("There are " + fmt.Sprint(totFiles) + " cache items.")

			selectedCacheItems = append(selectedCacheItems, cacheLocation)
			// if we have no arguments and the unused flag is set, we read the last build queries and clear the cache items that were not used
		} else if len(args) == 0 && unusedOption {
			unusedCacheItems, err := cm.GetUnusedCacheHashes()
			if err != nil {
				return utils.ErrorExit("Failed to get unused cache items.", err)
			}

			selectedCacheItems = append(selectedCacheItems, unusedCacheItems...)

			fmt.Println("Found " + fmt.Sprint(len(selectedCacheItems)) + " unused cache items.")
			// if we have one argument, we show the cache items for the query
		} else if len(args) == 1 {
			dirPath := cacheLocation + cache.Hash(args[0])

			files, err := os.ReadDir(dirPath)
			if err != nil {
				return utils.ErrorExit("Failed to read directory: ", err)
			}

			if len(files) > 1 {
				fmt.Println(args[0] + " represents a parameterized query with " + fmt.Sprint(len(files)) + " cache items.")
			} else {
				printFileContents(dirPath + "/" + files[0].Name())
			}

			selectedCacheItems = append(selectedCacheItems, dirPath)
			// if we have two or more arguments, we show the cache item for the query with the parameter
		} else if len(args) >= 2 {

			sparqlBytes, err := os.ReadFile("queries/" + args[0])
			if err != nil {
				return utils.ErrorExit("Failed to remove find query file.", err)
			}

			// for the second and all following arguments
			queryParameters := args[1:]
			query := string(sparqlBytes)
			for _, parameter := range queryParameters {
				argument := cast.ToString(parameter)
				query = strings.Replace(query, "{{.}}", argument, 1)
			}

			filePath := cacheLocation + cache.Hash(args[0]) + "/" + cache.Hash(query) + ".json"
			selectedCacheItems = append(selectedCacheItems, filePath)

			printFileContents((filePath))
		}

		if invalidateCacheOption {
			for _, item := range selectedCacheItems {
				fmt.Println("Removing: " + item)
				if err := os.RemoveAll(item); err != nil {
					return utils.ErrorExit("Failed to remove the cache file.", err)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.Flags().BoolVarP(&invalidateCacheOption, "invalidate", "i", false, "Removes/clears the specified parts of the query cache.")
	cacheCmd.Flags().BoolVarP(&unusedOption, "unused", "u", false, "Returns cache items not used in the last build.")
	cacheCmd.Flags().StringVarP(&snowmanPath, "snowman-directory", "d", ".snowman", "Sets the snowman directory to use.")
}
