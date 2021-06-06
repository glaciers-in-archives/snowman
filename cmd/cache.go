package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

var invalidateCacheOption bool

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
	Short: "Invalidates cache",
	Long:  `Removes all or specified parts of the query cache. Provide no argument to clear all cache. To clear the cache for a particular query provide an argument with the name of the query. To clear the cache for a parameterized query provide a second argument with its parameter value.`,
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if invalidateCacheOption {
				err := os.RemoveAll(cache.CacheLocation)
				if err != nil {
					return utils.ErrorExit("Failed to remove directory.", err)
				}
				return nil
			}

			fmt.Println("No arguments or flags given.")
		} else if len(args) == 1 {
			dirPath := cache.CacheLocation + cache.Hash(args[0])

			if invalidateCacheOption {
				err := os.RemoveAll(dirPath)
				if err != nil {
					return utils.ErrorExit("Failed to remove directory.", err)
				}
				return nil
			}

			files, err := os.ReadDir(dirPath)
			if err != nil {
				return utils.ErrorExit("Failed to read directory: ", err)
			}

			if len(files) > 1 {
				fmt.Println(args[0] + " represents a parameterized query with " + fmt.Sprint(len(files)) + " cache items.")
				return nil
			}

			return printFileContents(dirPath + "/" + files[0].Name())
		} else if len(args) == 2 {
			filePath := cache.CacheLocation + cache.Hash(args[0]) + "/" + cache.Hash(args[1]) + ".json"
			if invalidateCacheOption {
				if err := os.Remove(filePath); err != nil {
					return utils.ErrorExit("Failed to remove the cache file.", err)
				}
			}

			return printFileContents((filePath))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	buildCmd.Flags().BoolVarP(&invalidateCacheOption, "invalidate", "i", false, "Removes/clears specified parts of the query cache.")
}
