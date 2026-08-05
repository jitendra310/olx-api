package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	DatabaseUrl string
}

func MustLoad() Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		panic("PORT is required")
	}

	env := os.Getenv("ENV")
	if env == "" {
		panic("ENV is required")
	}

	DbUrl := os.Getenv("DATABASE_URL")
	if DbUrl == "" {
		panic("DATABASE_URL is required")
	}

	return Config{
		Port:        port,
		Env:         env,
		DatabaseUrl: DbUrl,
	}
}
