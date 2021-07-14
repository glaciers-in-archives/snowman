package utils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ErrorExit(message string, err error) error {
	return errors.New(message + " Error: " + err.Error())
}

func CopyDir(from string, to string) error {
	// This does not include checking if the "from" directory exists
	err := filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			newPath := strings.Replace(path, from+"/", to+"/", 1)
			if err := os.MkdirAll(filepath.Dir(newPath), 0770); err != nil {
				return err
			}

			original, err := os.Open(path)
			if err != nil {
				return err
			}
			defer original.Close()

			new, err := os.Create(newPath)
			if err != nil {
				return err
			}
			defer new.Close()

			_, err = io.Copy(new, original)
			if err != nil {
				return err
			}
		}
		return err
	})
	return err
}

func Join(sep string, strs ...string) string {
	return strings.Join(strs, sep)
}
