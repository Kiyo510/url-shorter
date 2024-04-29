package config

import (
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	User string
	Host string
	Name string
	Pass string
}

var PostgresConf DBConfig

func LoadDBConfig() {
	config := DBConfig{
		User: os.Getenv("DB_USER"),
		Host: os.Getenv("DB_HOST"),
		Name: os.Getenv("DB_NAME"),
		Pass: os.Getenv("DB_PASS"),
	}

	if config.User == "" || config.Host == "" || config.Name == "" || config.Pass == "" {
		log.Fatal("Missing required db environment variables")
	}

	PostgresConf = config
}
