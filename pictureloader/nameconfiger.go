package main

import (
	"errors"
	"fmt"
	"os"
	"path"
)

const (
	nameLimit        = 10000
	nameErrorMessage = "No valid name available:\n"
)

//getFileNameFromPath - return destination path with unique file name.
//args:
// configuration *argsConfig - program args entered by user
// pathValue string - may be url or local path, depends on needs
// isURL bool - marker, true - if pathValue is url, false - if not
func getFileNameFromPath(configuration *argsConfig, pathValue string, isURL bool) (string, error) {

	if isURL {
		pathValue = path.Join(configuration.destDir, path.Base(pathValue))
	} else {
		pathValue = path.Join(configuration.destAvatarDir, path.Base(pathValue))
	}

	if _, err := os.Stat(pathValue); os.IsNotExist(err) {
		return pathValue, nil
	}

	for i := 0; i < nameLimit; i++ {

		res := fmt.Sprintf("%s_(%d)", pathValue, i)
		if _, err := os.Stat(res); os.IsNotExist(err) {
			return res, nil
		}
	}

	return "", errors.New(nameErrorMessage + pathValue)
}
