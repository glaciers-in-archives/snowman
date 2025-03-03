package resources_cache_cmd

import (
	"github.com/spf13/cobra"
)

var unusedOption bool
var snowmanPath string

// cahceCmd represents the cache resources command
var ResourcesCacheCmd = &cobra.Command{
	Use:   "resources",
	Short: "Manage the Snowman resources cache.",
}

func init() {
	ResourcesCacheCmd.AddCommand(resourcesCacheInspectCmd)
	ResourcesCacheCmd.AddCommand(resourcesCacheClearCmd)

	ResourcesCacheCmd.PersistentFlags().StringVarP(&snowmanPath, "snowman-directory", "d", ".snowman", "Sets the snowman directory to use.")
	ResourcesCacheCmd.PersistentFlags().BoolVarP(&unusedOption, "unused", "u", false, "Returns cache items not used in the last build.")
}
