package main

import (
	"errors"
	"regexp"
	"strconv"
	"unicode"
)

const (
	pathPattern            = `(^(\/?[^<>:"\/\\|?*]+)+\/?$)|(^([a-zA-Z]\:\\)?(\\?[^<>:"\/\\|?*]+)+\\?$)` //pattern to check if local path value is correct.
	pathErrMessage         = "Incorrect path value:\n"
	fileNotExistErrMessage = "File does not exist:\n"
	uintErrMessage         = "The value must be an integer > = 1. Incorrect value:\n"
)

//pathValidator - check if local path is correct.
//Return:
// path value and nil - if correct.
// "" and err value - if not.
func pathValidator(path string) (string, error) {
	matched, err := regexp.MatchString(pathPattern, path)
	logErr(err)

	if matched {
		return path, nil
	}

	return "", errors.New(pathErrMessage + path)
}

//uintValidator - check if string can be interpreted as unsigned int.
//Return:
// interpreted value and nil - if may and value != 0.
// 0 and error value - if not.
func uintValidator(value string) (uint, error) {

	for _, r := range value {
		if !unicode.IsDigit(r) {
			return 0, errors.New(uintErrMessage + value)
		}
	}

	res, err := strconv.ParseUint(value, 0, 64)
	if err != nil || res < 1 {
		return 0, errors.New(uintErrMessage + value)
	}

	return uint(res), nil

}
