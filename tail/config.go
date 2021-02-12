package main

import (
	"fmt"

	"github.com/hpcloud/tail"
)

type config struct {
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
}

func (config *config) String() string {
	res := ""
	res = fmt.Sprintf("%vpath: %v;\nregexp: %v;\nuserPrefix: %v;\nn: %v;\n", res, config.path, config.regex, config.userPrefix, config.n)
	res = fmt.Sprintf("%vmessage prefix: %v\n", res, config.messagePrefix)
	res = fmt.Sprintf("%vlines amount to start read with: %v\n", res, config.n)
	res = fmt.Sprintf("%visFilepath: %v\n", res, config.isFilepath)
	res = fmt.Sprintf("%vLocation: %v\n", res, config.hpcloudTailCfg.Location)
	res = fmt.Sprintf("%vReopen: %v\n", res, config.hpcloudTailCfg.ReOpen)
	res = fmt.Sprintf("%vFollow: %v\n", res, config.hpcloudTailCfg.Follow)
	res = fmt.Sprintf("%vLogger hpcloud: %v", res, config.hpcloudTailCfg.Logger)
	return res
}
