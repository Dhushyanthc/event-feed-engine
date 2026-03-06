package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string
	Port string
	RedisUrl string
	DatabaseUrl string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		AppEnv: os.Getenv("APP_ENV"),
		Port : os.Getenv("PORT"),
		RedisUrl: os.Getenv("REDIS_URL"),
		DatabaseUrl: os.Getenv("DATABASE_URL"),
	}
	return cfg, nil
}