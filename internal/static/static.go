package static

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glaciers-in-archives/snowman/internal/utils"
)

func ClearStatic() error {
	lastStaticFiles, err := utils.ReadLineSeperatedFile(".snowman/static_history.txt")
	if err != nil {
		return err
	}

	for _, path := range lastStaticFiles {
		fmt.Println("Removing: " + path)
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return err
}

func CopyIn() error {
	var writtenFiles []string
	// This does not include checking if the "from" directory exists
	err := filepath.Walk("static", func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			newPath := strings.Replace(path, "static/", "site/", 1)
			if err := os.MkdirAll(filepath.Dir(newPath), 0770); err != nil {
				return err
			}

			err := utils.CopyFile(path, newPath)
			if err != nil {
				return err
			}
			writtenFiles = append(writtenFiles, newPath)
		}
		return err
	})

	err = utils.WriteLineSeperatedFile(writtenFiles, ".snowman/static_history.txt")
	return err
}
