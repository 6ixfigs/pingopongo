package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBConn     string
	BotToken   string
}

var (
	instance *Config
	once     sync.Once
)

func Get() (*Config, error) {
	var err error
	once.Do(func() {
		if loadErr := godotenv.Load(); err != nil {
			err = loadErr
			return
		}

		instance = &Config{
			ServerPort: os.Getenv("SERVER_PORT"),
			DBConn: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				os.Getenv("DB_USER"),
				os.Getenv("DB_PASSWORD"),
				os.Getenv("DB_HOST"),
				os.Getenv("DB_PORT"),
				os.Getenv("DB_NAME"),
			),
			BotToken: os.Getenv("BOT_TOKEN"),
		}
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}
