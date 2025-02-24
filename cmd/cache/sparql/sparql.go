package sparql_cache_cmd

import (
	"github.com/spf13/cobra"
)

var unusedOption bool
var snowmanPath string

// cahceCmd represents the cache sparql command
var SparqlCacheCmd = &cobra.Command{
	Use:   "sparql",
	Short: "Manage the Snowman SPARQL cache.",
}

func init() {
	SparqlCacheCmd.AddCommand(sparqlCacheInspectCmd)
	SparqlCacheCmd.AddCommand(sparqlCacheClearCmd)

	SparqlCacheCmd.PersistentFlags().StringVarP(&snowmanPath, "snowman-directory", "d", ".snowman", "Sets the snowman directory to use.")
	SparqlCacheCmd.PersistentFlags().BoolVarP(&unusedOption, "unused", "u", false, "Returns cache items not used in the last build.")
}
