package main

import (
	"flag"

	"github.com/snxq/rkeinstaller"
)

var (
	configpath = flag.String("conf", "config.yaml", "config file")
)

func main() {
	flag.Parse()

	var errhandle = rkeinstaller.ErrHandle

	conf, err := rkeinstaller.LoadConfigFromFile(*configpath)
	errhandle("load config from file failed", err)
	errhandle("check configs failed", conf.Check())

	errhandle("install rke failed",
		rkeinstaller.NewInstaller(conf).Start())
}
