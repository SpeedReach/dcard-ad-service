package infra

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	PostgresURI   string
	RedisURI      string
	AutoMigration bool
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
	log.Print("AUTO_MIGRATION ", config.AutoMigration)
	return config
}

func loadFromOS() Config {
	return Config{
		PostgresURI:   os.Getenv("POSTGRES_URI"),
		RedisURI:      os.Getenv("REDIS_URI"),
		AutoMigration: os.Getenv("AUTO_MIGRATION") == "true",
	}
}

// todo: implement this
func loadFromInfisical(serviceToken string) Config {
	panic("Unimplemented")
}
