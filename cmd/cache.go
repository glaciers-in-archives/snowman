package cmd

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

var invalidateCacheOption bool
var unusedOption bool

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

	b, err := ioutil.ReadAll(file)
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
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedCacheItems []string

		if len(args) == 0 && !unusedOption {
			totFiles, err := utils.CountFilesRecursive(cache.CacheLocation)
			if err != nil {
				return utils.ErrorExit("Failed to retrive cache info.", err)
			}

			fmt.Println("There are " + fmt.Sprint(totFiles) + " cache items.")

			selectedCacheItems = append(selectedCacheItems, cache.CacheLocation)
		} else if len(args) == 0 && unusedOption {
			usedItems, err := utils.ReadLineSeperatedFile(".snowman/last_build_queries.txt")
			if err != nil {
				return utils.ErrorExit("Failed to read last unused cache items: ", err)
			}

			err = fs.WalkDir(os.DirFS("."), cache.CacheLocation, func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				pathAsCacheItem := strings.Replace(strings.Replace(path, ".json", "", 1), cache.CacheLocation, "", 1)
				isUsed := false
				for _, used := range usedItems {
					if pathAsCacheItem == used || strings.HasPrefix(used, pathAsCacheItem) {
						isUsed = true
					}
				}

				if !isUsed {
					selectedCacheItems = append(selectedCacheItems, path)
				}
				return nil
			})

			fmt.Println("Found " + fmt.Sprint(len(selectedCacheItems)) + " unused cache items.")
		} else if len(args) == 1 {
			dirPath := cache.CacheLocation + cache.Hash(args[0])

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
		} else if len(args) == 2 {

			sparqlBytes, err := ioutil.ReadFile("queries/" + args[0])
			if err != nil {
				return utils.ErrorExit("Failed to remove find query file.", err)
			}

			queryString := strings.Replace(string(sparqlBytes), "{{.}}", args[1], 1)

			filePath := cache.CacheLocation + cache.Hash(args[0]) + "/" + cache.Hash(queryString) + ".json"
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
}
