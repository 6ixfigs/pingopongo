package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBConn     string
}

func Get() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		ServerPort: os.Getenv("SERVER_PORT"),
		DBConn: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		),
	}, nil
}
