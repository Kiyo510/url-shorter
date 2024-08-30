package config

import (
	"log"
	"os"
)

type AppConfig struct {
	BaseUrl            string
	NewRelicLicenseKey string
	AppName            string
}

var AppConf AppConfig

func LoadAppConfig() {
	conf := AppConfig{
		BaseUrl:            os.Getenv("BASE_URL"),
		NewRelicLicenseKey: os.Getenv("NEW_RELIC_LICENSE_KEY"),
		AppName:            os.Getenv("APP_NAME"),
	}

	if conf.BaseUrl == "" {
		log.Fatal("Missing required environment variable BASE_URL")
	}

	if conf.NewRelicLicenseKey == "" {
		log.Fatal("Missing required environment variable NEW_RELIC_LICENSE_KEY")
	}

	if conf.AppName == "" {
		log.Fatal("Missing required environment variable APP_NAME")
	}

	AppConf = conf
}
