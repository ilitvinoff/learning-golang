package main

import (
	"fmt"

	"github.com/hpcloud/tail"
)

type config struct {
	tailConfig        *tail.Config
	path              string
	regex             string
	prefix            string
	n                 int
	isFilepath        bool
	readFromBeginning bool
}

func (config *config) String() string {
	res := ""
	res = fmt.Sprintf("%vpath: %v;\nregexp: %v;\nprefix: %v;\nn: %v;\n", res, config.path, config.regex, config.prefix, config.n)
	//res = fmt.Sprintf("%vLocation: %v\n", res, config.tailConfig.Location)
	res = fmt.Sprintf("%vReopen: %v\n", res, config.tailConfig.ReOpen)
	res = fmt.Sprintf("%vFollow: %v\n", res, config.tailConfig.Follow)
	res = fmt.Sprint(res, "-------------------------------------------------------\n")
	return res
}
