package config

import (
	"log"
	"os"
)

type RedisConfig struct {
	Host string
	Port string
}

var RedisConf RedisConfig

func LoadRedisConfig() {
	config := RedisConfig{
		Host: os.Getenv("REDIS_HOST"),
		Port: os.Getenv("REDIS_PORT"),
	}

	if config.Host == "" || config.Port == "" {
		log.Fatal("Missing required redis environment variables")
	}

	RedisConf = config
}
