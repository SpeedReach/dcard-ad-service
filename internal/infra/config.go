package infra

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	PostgresURI string
	RedisURI    string
}

func LoadConfig() Config {
	_ = godotenv.Load()
	config := loadFromOS()
	if config.PostgresURI == "" {
		panic("Missing PostgresURI")
	}
	if config.RedisURI == "" {
		panic("Missing RedisURI")
	}
	return config
}

func loadFromOS() Config {
	return Config{
		PostgresURI: os.Getenv("POSTGRES_URI"),
		RedisURI:    os.Getenv("REDIS_URI"),
	}
}

// todo: implement this
func loadFromInfisical(serviceToken string) Config {
	panic("Unimplemented")
}
