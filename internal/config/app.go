package config

import (
	"os"
)

var AppConfig struct {
	BaseUrl string
}

func LoadAppConfig() {
	AppConfig.BaseUrl = os.Getenv("BASE_URL")
}
