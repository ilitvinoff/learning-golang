package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	nDefaultValue         = 10
	watchPollDelayDefault = 100
)

//UserWatchPollDellay - value of the watch poll delay set by user
var UserWatchPollDellay time.Duration

type pathFlag []*config

type regexFlag []*config

type configFlag []*config

func (p *pathFlag) Set(value string) error {
	args := strings.Split(value, ";")
	argsLength := len(args)

	if argsLength == 1 {
		cfg := getDefaultConfig()
		cfg.path = args[0]
		cfg.isFilepath = true
		*p = append(*p, cfg)
		return nil
	}

	if argsLength%2 != 0 {
		return fmt.Errorf("not enough arguments set with flag '-p'")
	}

	for i := 0; i < argsLength; {
		cfg := getDefaultConfig()
		cfg.path = args[i]
		cfg.userPrefix = args[i+1]
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
	args := strings.Split(value, ";")
	argsLength := len(args)

	if argsLength < 2 {
		return fmt.Errorf("not enough arguments set with '-r'")
	}

	if argsLength == 2 {
		cfg := getDefaultConfig()
		cfg.path = args[0]
		cfg.regex = args[1]
		*r = append(*r, cfg)
		return nil
	}

	if argsLength%3 != 0 {
		return fmt.Errorf("not enough arguments set with flag '-p'")
	}

	for i := 0; i < argsLength; {
		cfg := getDefaultConfig()
		cfg.path = args[i]
		cfg.regex = args[i+1]
		cfg.userPrefix = args[i+2]
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
	args := strings.Split(value, ";")
	argsLength := len(args)

	if argsLength%4 != 0 {
		return fmt.Errorf("not enough arguments set with flag '-c'")
	}

	for i := 0; i < argsLength; {
		var err error
		offset := nDefaultValue

		if args[i+3] != "" {

			offset, err = strconv.Atoi(args[i+3])
			if err != nil {
				return fmt.Errorf("wrong value set for amount of strings to tail from: %v", err)
			}
		}

		cfg := getDefaultConfig()
		cfg.path = args[i]
		cfg.regex = args[i+1]
		cfg.userPrefix = args[i+2]
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
	var watchPollDelayInt int

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tail -flag \"configuration string\" -flag1 \"configuration string\"...\n	Print the last 10 lines of each FILE to standard output.\n\nThe basic functionality is the same as the standard tail utility started with '-F' flag.\n\nAdded functionality:\n	1. You can \"tail\" multiple files at the same time. This way, you can define a \"prefix\" for each file, \n	which will be printed before the line from the corresponding file.\n	2. If file - doesn't exist -> wait for it to appear\n	3. If the file has been deleted / moved -> wait for a new one to appear. When new one appeared, tailing starts from the begining of the file.\n	4. You may select directory and define the regular expression(for file name).\n	This way, 'tailer' will keep track of the last file that came up that matches the regular expression.\n\nWarning!!!!\nYou need to describe parameters of each flag with 1 stringline. So, for example you need to tail 2 files, then both files\nyou'll describe in one string, using semicolon, as delimiter.\n\n	Example:\n	./tail -p \"filepath1;prefix for output from file1;filepath2;prefix for output from file2...\"\n\nWarning!!!\nDo not use semicolon at the end of argument line.\n\n Available flags:\n")

		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "	-%v\n%v\n", f.Name, f.Usage) // f.Name, f.Value
		})
	}

	flag.Var(&pathFlag, "p", "File's path to tail.If you want to specify 1 file path, 1 argument is enough - the file path.\nYou may specify prefix to printout it before textline from file, using semicolon as delimiter. Example: tail -p \"foo/bar/file;prefix\"\nIf you want to specify more then 1 file, you need to define both parameters(file's path;prefix) for each file in one string.\nExample: tail -p \"foo/bar/file1;prefix1;foo/bar/file2;prefix2;foo/bar/file3;prefix3...\"\nWarning!!!! Do not use semicolon at the end of argument line.\n")
	flag.Var(&regexFlag, "r", "Filename regular expression pattern. If you want to specify 1 file to tail, 2 arguments are enough - \"pathToDirectory;regex\".\nIf you want to specify more then 1 file, you need to define 3 arguments for each file: \nExample: tail -r \"foo/bar/file1;regexForFile1;prefix1;foo/bar/file2;regexForFile2;prefix2;foo/bar/file3;regexForFile3;prefix3...\"\nWarning!!!! Do not use semicolon at the end of argument line.\n ")
	flag.Var(&configFlag, "c", "Set's full config for each file to tail. All files specify in 1 string, using semicolon as delimiter. You need to define such parameters for each file:\n path - path to file or directory, depends on your will\n regex - regular expression for filename(may be empty string, if u set path - as path to file. empty string means: path;;prefix;n)\n prefix - prefix to printout before line from file\n n: output the last 'n' lines,may be empty string(Must be integer if present!!!), \nExample: tail -c \"foo/bar/file1;regexForFile1;prefix1;someinteger;foo/bar/file2;regexForFile2;prefix2;someinteger;...\"\nWarning!!!! Do not use semicolon at the end of argument line.\n")
	flag.IntVar(&n, "n", nDefaultValue, "when 'n' is set, it defines to all files, where parameter 'n' has default value(default - 10) . 'n' represent amount of string to tail from.\nWarning!!!! Do not use semicolon at the end of argument line.\n")
	flag.IntVar(&watchPollDelayInt, "delay", watchPollDelayDefault, "Set watch poll delay value (milliseconds). If not set by the user or set to zero, then initialized with a default value of 100 milliseconds.")
	flag.Parse()

	res = appendAllConfigsFromFlags(pathFlag, regexFlag, configFlag)

	if n != nDefaultValue {
		for _, el := range res {
			if el.n == nDefaultValue {
				el.n = n
			}
		}
	}

	UserWatchPollDellay = time.Duration(watchPollDelayDefault) * time.Millisecond

	if watchPollDelayInt != watchPollDelayDefault && watchPollDelayInt != 0 {
		UserWatchPollDellay = time.Duration(watchPollDelayInt) * time.Millisecond
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
