package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	// Config -.
	Config struct {
		App      App
		Log      Log
		PG       PG
		Gemini   Gemini
		Telegram Telegram
		JWT       JWT
		Redis     Redis
		ChatGpt   ChatGpt
	}

	JWT struct {
		Secret string `env-required:"true" env:"JWT_SECRET"`
	}
	// App -.
	App struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// Log -.
	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// PG -.
	PG struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	Gemini struct {
		APIKey string `env:"GEMINI_API_KEY,required"`
	}
	ChatGpt struct {
		APIKey string `env:"CHAT_GPT,required"`
	}
	
	Telegram struct {
		Token string `env:"TELEGRAM_BOT_TOKEN,required"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST,required"`
		Port     string `env:"REDIS_PORT,required"`
		
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	fmt.Println("config: ", cfg)

	return cfg, nil
}
