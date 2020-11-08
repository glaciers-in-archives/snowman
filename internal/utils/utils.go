package utils

import "errors"

func ErrorExit(message string, err error) error {
	return errors.New(message + " Error: " + err.Error())
}
