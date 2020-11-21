package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/hpcloud/tail"
)

const (
	nDefaultValue = 10
)

type pathFlag []*config

type regexFlag []*config

type configFlag []*config

func defaultParams() *config {
	return &config{&tail.Config{Follow: true}, "", "", "", nDefaultValue, false, false}
}

func (p *pathFlag) Set(value string) error {
	if value == "" {
		return fmt.Errorf("empty path set with flag '-p'")
	}

	args := strings.Split(value, ";")
	argsLength := len(args)

	if argsLength == 1 {
		cfg := defaultParams()
		cfg.path = args[0]
		cfg.isFilepath = true
		*p = append(*p, cfg)
		return nil
	}

	if argsLength%2 != 0 {
		return fmt.Errorf("not enough arguments set with flag '-p'")
	}

	for i := 0; i < argsLength; {
		cfg := defaultParams()
		cfg.path = args[i]
		cfg.prefix = args[i+1]
		cfg.isFilepath = true
		*p = append(*p, cfg)
		i += 2
	}

	return nil
}

func (p *pathFlag) String() string {
	res := ""
	for el := range *p {
		res = fmt.Sprintf(res, el, "\n")
	}
	return res
}

func (r *regexFlag) Set(value string) error {
	if value == "" {
		return fmt.Errorf("empty path set with flag '-r'")
	}

	args := strings.Split(value, ";")
	argsLength := len(args)

	if argsLength < 2 {
		return fmt.Errorf("not enough arguments set with '-r'")
	}

	if argsLength == 2 {
		cfg := defaultParams()
		cfg.path = args[0]
		cfg.regex = args[1]
		*r = append(*r, cfg)
		return nil
	}

	if argsLength%3 != 0 {
		return fmt.Errorf("not enough arguments set with flag '-p'")
	}

	for i := 0; i < argsLength; {
		cfg := defaultParams()
		cfg.path = args[i]
		cfg.regex = args[i+1]
		cfg.prefix = args[i+2]
		*r = append(*r, cfg)
		i += 3
	}

	return nil
}

func (r *regexFlag) String() string {
	res := ""
	for el := range *r {
		res = fmt.Sprintf(res, el, "\n")
	}
	return res
}

func (c *configFlag) Set(value string) error {
	if value == "" {
		return fmt.Errorf("empty path set with flag '-c'")
	}

	args := strings.Split(value, ";")
	argsLength := len(args)

	if argsLength%4 != 0 {
		return fmt.Errorf("not enough arguments set with flag '-c'")
	}

	for i := 0; i < argsLength; {
		offset, err := strconv.Atoi(args[i+3])
		if err != nil {
			return fmt.Errorf("wrong value set for amount of strings to tail from: %v", err)
		}

		cfg := defaultParams()
		cfg.path = args[i]
		cfg.regex = args[i+1]
		cfg.prefix = args[i+2]
		cfg.n = offset
		if cfg.regex == "" {
			cfg.isFilepath = true
		}

		*c = append(*c, cfg)
		i += 4
	}

	return nil
}

func (c *configFlag) String() string {
	res := ""
	for el := range *c {
		res = fmt.Sprintf(res, el, "\n")
	}
	return res
}

func getConfigFromFlags() []*config {
	var res []*config
	var pathFlag pathFlag
	var regexFlag regexFlag
	var configFlag configFlag
	var n int

	flag.Var(&pathFlag, "p", "File's path to tail.If you want to specify 1 file path, 1 argument is enough - the file path.\nYou may specify prefix to printout it before textline from file, using semicolon as delimiter. Example: tail -p \"foo/bar/file;prefix\"\nIf you want to specify more then 1 file, you need to define both parameters(file's path;prefix) for each file in one string.\nExample: tail -p \"foo/bar/file1;prefix1;foo/bar/file2;prefix2;foo/bar/file3;prefix3...\"")
	flag.Var(&regexFlag, "r", "Filename regular expression pattern. If you want to specify 1 file to tail, 2 arguments are enough - \"pathToDirectory;regex\".\nIf you want to specify more then 1 file, you need to define 3 arguments for each file: \nExample: tail -r \"foo/bar/file1;regexForFile1;prefix1;foo/bar/file2;regexForFile2;prefix2;foo/bar/file3;regexForFile3;prefix3...\" ")
	flag.Var(&configFlag, "c", "Set's full config for each file to tail. All files specify in 1 string, using semicolon as delimiter. You need to define such parameters for each file:\n path - path to file or directory, depends on your will\n regex - regular expression for filename(may be empty string, if u set path - as path to file. empty string means: path;;prefix;n)\n prefix - prefix to printout before line from file\n n: output the last 'n' lines,may be empty string(Must be integer if present!!!), \nExample: tail -c \"foo/bar/file1;regexForFile1;prefix1;someinteger;foo/bar/file2;regexForFile2;prefix2;someinteger;...\"")
	flag.IntVar(&n, "n", nDefaultValue, "when 'n' is set, it defines to all files, where parameter 'n' has default value(default - 10) . 'n' represent amount of string to tail from.")
	flag.Parse()

	res = appendAllConfigsFromFlags(pathFlag, regexFlag, configFlag)

	if n != nDefaultValue {
		for _, el := range res {
			if el.n == nDefaultValue {
				el.n = n
			}
		}
	}

	return res
}

func appendAllConfigsFromFlags(pathFlag pathFlag, regexFlag regexFlag, configFlag configFlag) []*config {
	var res []*config
	if pathFlag != nil {
		for _, el := range pathFlag {
			res = append(res, el)
		}
	}

	if regexFlag != nil {
		for _, el := range regexFlag {
			res = append(res, el)
		}
	}

	if configFlag != nil {
		for _, el := range configFlag {
			res = append(res, el)
		}
	}
	return res
}
