package config

import (
	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Conf struct {
	Server ConfServer
	DB     ConfigDB
	FR     ConfFrontend
}

type ConfServer struct {
	Port         int           `env:"SERVER_PORT,required"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
	Debug        bool          `env:"SERVER_DEBUG,required"`
}

type ConfigDB struct {
	Host        string `env:"DB_HOST,required"`
	Port        int    `env:"DB_PORT,required"`
	Username    string `env:"DB_USER,required"`
	Password    string `env:"DB_PASS,required"`
	DBName      string `env:"DB_NAME,required"`
	Debug       bool   `env:"DB_DEBUG,required"`
	AutoMigrate bool   `env:"DB_AUTO_MIGRATE,required"`
}

type ConfFrontend struct {
	Port int    `env:"FRONTEND_PORT,required"`
	Host string `env:"FRONTEND_HOST,required"`
}

func New() *Conf {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
