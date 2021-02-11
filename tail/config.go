package main

import (
	"fmt"

	"github.com/hpcloud/tail"
)

type config struct {
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
}

func (config *config) String() string {
	res := ""
	res = fmt.Sprintf("%vpath: %v;\nregexp: %v;\nuserPrefix: %v;\nn: %v;\n", res, config.path, config.regex, config.userPrefix, config.n)
	res = fmt.Sprintf("%vLocation: %v\n", res, config.hpcloudTailCfg.Location)
	res = fmt.Sprintf("%vReopen: %v\n", res, config.hpcloudTailCfg.ReOpen)
	res = fmt.Sprintf("%vFollow: %v\n", res, config.hpcloudTailCfg.Follow)
	res = fmt.Sprint(res, "-------------------------------------------------------\n")
	return res
}
