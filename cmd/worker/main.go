package main

import (
	"log"

	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
	"pulseroad/internal/pkg/rabbitmq"
)

func StartWorker(cfg *config.Config) {
	db, err := database.Init(&cfg.MySQL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer database.Close(db)

	if err := rabbitmq.ValidateURL(cfg.RabbitMQ.URL); err != nil {
		log.Fatalf("invalid rabbitmq config: %v", err)
	}

	log.Printf("[%s] Worker process started successfully (env=%s)", cfg.App.Name, cfg.App.Env)
	log.Printf("RabbitMQ configured at %s; no background consumers are registered yet", cfg.RabbitMQ.URL)
}

func main() {
	// 加载配置
	cfg, err := config.Load("internal/pkg/config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	StartWorker(cfg)
}
