package config

import (
	"cc.tim/client/model"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

var Config model.Config

func Init(addr string) {
	configFile, err := ioutil.ReadFile(addr)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(configFile, &Config)

	Config.Avatar.MaxSize = 1024 * Config.Avatar.MaxSize
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}
