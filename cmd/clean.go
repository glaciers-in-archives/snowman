package cmd

import (
	"fmt"
	"os"

	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes the site directory.",
	Long:  `Tries to remove the site directory and all it contents so that Snowman can build a new site in its place.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if _, err := os.Stat("site"); err != nil {
			fmt.Println("The site directory does not exists.")
			return nil
		}

		if err := os.RemoveAll("site"); err != nil {
			return utils.ErrorExit("Failed to remove the site directory.", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
