package main

import (
	"log"

	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
)

//
func StartWorker(cfg *config.Config) {
	db, err := database.Init(&cfg.MySQL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer database.Close(db)
	_ = db

	log.Printf("[%s] Worker process started successfully (env=%s)", cfg.App.Name, cfg.App.Env)
}

func main() {
	// 加载配置
	cfg, err := config.Load("internal/pkg/config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	StartWorker(cfg)
}
