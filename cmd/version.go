package cmd

import (
	"fmt"
	"runtime"

	"github.com/glaciers-in-archives/snowman/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the Snowman version",
	Long:  `Prints the Snowman version and additional build information.`,
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Snowman " + version.CurrentVersion.String() + " " + runtime.GOOS + "/" + runtime.GOARCH)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
