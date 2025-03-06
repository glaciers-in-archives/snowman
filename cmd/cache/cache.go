package cache_cmd

import (
	resources_cache_cmd "github.com/glaciers-in-archives/snowman/cmd/cache/resources"
	sparql_cache_cmd "github.com/glaciers-in-archives/snowman/cmd/cache/sparql"
	"github.com/spf13/cobra"
)

// cahceCmd represents the cache command
var CacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the Snowman cache.",
}

func init() {
	CacheCmd.AddCommand(sparql_cache_cmd.SparqlCacheCmd)
	CacheCmd.AddCommand(resources_cache_cmd.ResourcesCacheCmd)
}
