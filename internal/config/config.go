package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	DatabaseUrl string
	JwtKet      string
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

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		panic("JWT_KEY is required")
	}

	return Config{
		Port:        port,
		Env:         env,
		DatabaseUrl: DbUrl,
		JwtKet:      jwtKey,
	}
}
