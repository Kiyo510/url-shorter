package config

import (
	"log"
	"os"
)

type AppConfig struct {
	BaseUrl string
}

var AppConf AppConfig

func LoadAppConfig() {
	conf := AppConfig{
		BaseUrl: os.Getenv("BASE_URL"),
	}

	if conf.BaseUrl == "" {
		log.Fatal("Missing required environment variable BASE_URL")
	}

	AppConf = conf
}
