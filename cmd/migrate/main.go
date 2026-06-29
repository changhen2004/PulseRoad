package main

import (
	"log"

	_ "pulseroad/internal/auth"
	_ "pulseroad/internal/feedback"
	_ "pulseroad/internal/flagflow"
	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
	_ "pulseroad/internal/product"
	_ "pulseroad/internal/requirement"
	_ "pulseroad/internal/team"
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
