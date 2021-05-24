package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// CLI ARGUMENTS
var cache string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "snowman <command> [flags]",
	Short: "A static site generator for SPARQL backends. ",
	Long:  `Snowman is a CLI tool for creating websites from SPARQL queries.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	buildCmd.Flags().StringVarP(&cache, "cache", "c", "available", "Sets the cache strategy. \"available\" will use cached SPARQL responses when available and fallback to making queries. \"never\" will ignore existing cache and will not update or set new cache.")
}
