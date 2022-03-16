package cmd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/utils"
	"github.com/spf13/cobra"
)

var directory string

//go:embed scaffold/*
var content embed.FS

// serverCmd represents the server command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Generates a new project.",
	Long:  ``,
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		if _, err := os.Stat(directory); !os.IsNotExist(err) {
			fmt.Println("Directory already exists.")
			return nil
		}

		fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			newPath := strings.Replace(path, "scaffold/", directory+"/", 1)

			if err := os.MkdirAll(filepath.Dir(newPath), 0770); err != nil {
				utils.ErrorExit("Failed to generate directory: ", err)
			}

			out, err := os.Create(newPath)
			defer out.Close()
			if err != nil {
				utils.ErrorExit("Failed to create file: ", err)

			}

			in, err := embed.FS.Open(content, path)
			defer in.Close()
			if err != nil {
				utils.ErrorExit("Failed to open embeded filesystem: ", err)

			}

			_, err = io.Copy(out, in)
			if err != nil {
				utils.ErrorExit("Failed to copy file: ", err)

			}

			return nil
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&directory, "directory", "d", "my-new-project", "Address to which the server will bind.")
}
