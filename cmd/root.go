package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var timeit bool
var verbose bool

func printVerbose(message string) {
	if verbose {
		fmt.Println(message)
	}
}

func elapsed() func() {
	start := time.Now()
	return func() {
		if timeit {
			fmt.Println("Finished in " + time.Since(start).String())
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "snowman <command> [flags]",
	Short: "A static site generator for SPARQL backends. ",
	Long:  `Snowman is a CLI tool for creating websites from SPARQL queries.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer elapsed()()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&timeit, "timeit", "t", false, "")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Activate verbose output.")
}
