package cmd

import (
	"os"

	"github.com/glaciers-in-archives/snowman/internal/cache"
	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

// clearCahceCmd represents the clearcache command
var clearCahceCmd = &cobra.Command{
	Use:   "clearcache",
	Short: "Invalidates cache",
	Long:  `Removes all or specified parts of the query cache. Provide no argument to clear all cache. To clear the cache for a particular query provide an argument with the name of the query. To clear the cache for a dynamic query provide a second argument with its parameter value.`,
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if err := os.RemoveAll(cache.CacheLocation); err != nil {
				return utils.ErrorExit("Failed to remove the cache directory.", err)
			}
		} else if len(args) == 1 {
			if err := os.RemoveAll(cache.CacheLocation + cache.Hash(args[0])); err != nil {
				return utils.ErrorExit("Failed to remove the cache directory.", err)
			}
		} else {
			if err := os.Remove(cache.CacheLocation + cache.Hash(args[0]) + "/" + cache.Hash(args[1]) + ".json"); err != nil {
				return utils.ErrorExit("Failed to remove the cache file.", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clearCahceCmd)
}
