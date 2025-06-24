package main

import (
	"log"

	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
