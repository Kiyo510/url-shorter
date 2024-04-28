package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DBConfig struct {
	User string
	Host string
	Name string
	Pass string
	DB   *sqlx.DB
}

func LoadDBConfig() {
	DBConfig.User = os.Getenv("DB_USER")
	DBConfig.Host = os.Getenv("DB_HOST")
	DBConfig.Name = os.Getenv("DB_NAME")
	DBConfig.Pass = os.Getenv("DB_PASS")

	if DBConfig.User == "" || DBConfig.Host == "" || DBConfig.Name == "" || DBConfig.Pass == "" {
		log.Fatal("Missing required environment variables")
	}

	InitDB()
}

func InitDB() {
	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s", DBConfig.User, DBConfig.Pass, DBConfig.Host, DBConfig.Name, "disable")
	var err error
	DBConfig.DB, err = sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	}
}
