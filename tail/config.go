package main

import (
	"fmt"

	"github.com/hpcloud/tail"
)

type config struct {
<<<<<<< HEAD
	hpcloudTailCfg *tail.Config
	path           string
	regex          string
	userPrefix     string
	messagePrefix  string
	n              int
	isFilepath     bool
}

func getDefaultConfig() *config {
	return &config{&tail.Config{ReOpen: true, Follow: true, Logger: tail.DiscardingLogger, Location: &tail.SeekInfo{Offset: 0, Whence: 2}}, "", "", "", "", nDefaultValue, false}
=======
	hpcloudTailCfg    *tail.Config
	path              string
	regex             string
	userPrefix        string
	messagePrefix     string
	n                 int
	isFilepath        bool
	readFromBeginning bool
}

func getDefaultConfig() *config {
	return &config{&tail.Config{ReOpen: true, Follow: true, Logger: tail.DiscardingLogger}, "", "", "", "", nDefaultValue, false, false}
>>>>>>> a3fab59cdabc7435959b3fce756f4251b40be15b
}

func (config *config) String() string {
	res := ""
	res = fmt.Sprintf("%vpath: %v;\nregexp: %v;\nuserPrefix: %v;\nn: %v;\n", res, config.path, config.regex, config.userPrefix, config.n)
<<<<<<< HEAD
	res = fmt.Sprintf("%vmessage prefix: %v\n", res, config.messagePrefix)
	res = fmt.Sprintf("%vlines amount to start read with: %v\n", res, config.n)
	res = fmt.Sprintf("%visFilepath: %v\n", res, config.isFilepath)
	res = fmt.Sprintf("%vLocation: %v\n", res, config.hpcloudTailCfg.Location)
	res = fmt.Sprintf("%vReopen: %v\n", res, config.hpcloudTailCfg.ReOpen)
	res = fmt.Sprintf("%vFollow: %v\n", res, config.hpcloudTailCfg.Follow)
	res = fmt.Sprintf("%vLogger hpcloud: %v", res, config.hpcloudTailCfg.Logger)
=======
	res = fmt.Sprintf("%vLocation: %v\n", res, config.hpcloudTailCfg.Location)
	res = fmt.Sprintf("%vReopen: %v\n", res, config.hpcloudTailCfg.ReOpen)
	res = fmt.Sprintf("%vFollow: %v\n", res, config.hpcloudTailCfg.Follow)
	res = fmt.Sprint(res, "-------------------------------------------------------\n")
>>>>>>> a3fab59cdabc7435959b3fce756f4251b40be15b
	return res
}
