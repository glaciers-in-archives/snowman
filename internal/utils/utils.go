package utils

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ErrorExit(message string, err error) error {
	return errors.New(message + " Error: " + err.Error())
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

func WriteLineSeperatedFile(data []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	for i, value := range data {
		var line string = "\n" + value
		if i == 0 {
			line = value
		}
		_, err := writer.WriteString(line)
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

var illegalPath = regexp.MustCompile(`[\~\:\*\?\"\<\>\|]`)
var illegalNextToEachOther = regexp.MustCompile(`[\.\/]{2,}`)
var illegalStartAndEnd = regexp.MustCompile(`^[\./]|[\./]$`)

func ValidatePathSection(path string) error {
	// throw an error if the path contains illegal characters
	if illegalPath.MatchString(path) {
		return errors.New("Illegal characters in path: " + path)
	}

	// throw an error if the path contains . or / next to each other
	if illegalNextToEachOther.MatchString(path) {
		return errors.New("Illegal character combination in path: " + path)
	}

	// throw an error if the path starts or ends with . or /
	if illegalStartAndEnd.MatchString(path) {
		return errors.New("Illegal start or end in path: " + path)
	}

	// throw an error if the path is empty
	if path == "" {
		return errors.New("Path can't be empty")
	}

	return nil
}
