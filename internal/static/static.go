package static

import (
	"fmt"
	"io/fs"
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
	err := fs.WalkDir(os.DirFS("."), "static", func(path string, d fs.DirEntry, err error) error {
		info, _ := d.Info()
		if info.Mode().IsRegular() { // checks that its not ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice | ModeCharDevice | ModeIrregular
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
