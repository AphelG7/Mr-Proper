package configuration

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

type configInfo struct {
	Telegram telegramConfig
	Mongo    mongoConfig
}

type telegramConfig struct {
	Token string
	Debug bool
}

type mongoConfig struct {
	Url      string
	Database string
	Debug    bool
}

var Config = readConfig("mrproper.config")

func readConfig(configfile string) configInfo {
	_, err := os.Open(configfile)
	if err != nil {
		log.Fatal(err)
	}

	var config configInfo
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	return config
}
