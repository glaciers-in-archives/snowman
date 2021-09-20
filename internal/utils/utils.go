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

			err := CopyFile(path, newPath)
			if err != nil {
				return err
			}
		}
		return err
	})
	return err
}

func CopyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}
