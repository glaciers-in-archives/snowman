package utils

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
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

func WriteLineSeperatedFile(data map[string]bool, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	for value := range data {
		_, err := writer.WriteString(value + "\n")
		if err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func ReadLineSeperatedFile(path string) ([]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	values := strings.Split(string(bytes), "\n")
	return values, nil
}

func CountFilesRecursive(dir string) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, nil
	}

	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			count += 1

		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}
