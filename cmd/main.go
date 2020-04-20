package main

import (
	"flag"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/heqingy/TrashScheduler/fsm"
	"github.com/heqingy/TrashScheduler/svc"
	"gopkg.in/yaml.v2"
)

// Config
type Config struct {
	Users     []*fsm.User
	SMSConfig svc.SMSConfig
	FSMConfig svc.FSMConfig
}

func main() {
	var confPath string
	flag.StringVar(&confPath, "config", "./config.conf", "the path to the config file")
	flag.Parse()
	var c Config
	data, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalf(
			"failed to load config file %v: %v, please generate a file based on the following example:\n\n%v",
			confPath,
			err,
			yamlConfigTemplate,
		)
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Fatalf(
			"failed to unmarshal the config file: %v, please reset the config based on the following example:\n\n%v",
			err,
			yamlConfigTemplate,
		)
	}
	service := svc.NewService(fsm.UserListToMap(c.Users), svc.Config{
		Port:      8080,
		SMSConfig: c.SMSConfig,
		FSMConfig: c.FSMConfig,
	})
	service.Run()
}
