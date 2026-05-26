package main

import (
	"log"

	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
)

func main() {
	// 加载配置
	cfg, err := config.Load("internal/pkg/config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化数据库连接
	db, err := database.Init(&cfg.MySQL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer database.Close(db)

	// 运行自动迁移
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to run migration: %v", err)
	}

	log.Println("Migration completed successfully")
}
