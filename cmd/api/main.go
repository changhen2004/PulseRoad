package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/auth"
	"pulseroad/internal/feedback"
	"pulseroad/internal/flagflow"
	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
	"pulseroad/internal/pkg/logger"
	"pulseroad/internal/pkg/rabbitmq"
	"pulseroad/internal/pkg/redis"
	"pulseroad/internal/pkg/response"
	"pulseroad/internal/product"
	"pulseroad/internal/requirement"
	"pulseroad/internal/team"
)

// StartHttpServer 启动 HTTP 服务器。
func StartHttpServer(cfg *config.Config) {
	db, err := database.Init(&cfg.MySQL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer database.Close(db)

	redisClient, err := redis.Init(&cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis connection: %v", err)
		}
	}()

	rabbitClient, err := rabbitmq.DialWithRetry(context.Background(), cfg.RabbitMQ.URL, 30, time.Second)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer func() {
		if err := rabbitClient.Close(); err != nil {
			log.Printf("failed to close rabbitmq connection: %v", err)
		}
	}()

	r := gin.New() // 使用 Gin 的默认日志和恢复中间件
	r.Use(gin.Recovery())
	r.Use(logger.RequestLogger())

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 注册路由
	loginLimiter := auth.NewRedisLoginLimiter(redisClient, 5, 15*time.Minute)
	authService := auth.NewServiceWithLoginLimiter(auth.NewRepository(db), cfg.JWT.Secret, loginLimiter)
	auth.RegisterRoutes(r.Group("/api"), authService)
	teamService := team.NewService(team.NewRepository(db))
	team.RegisterRoutes(r.Group("/api"), authService, teamService)
	productService := product.NewService(product.NewRepository(db), teamService)
	product.RegisterRoutes(r.Group("/api"), authService, productService)
	feedbackPublisher := feedback.NewRabbitMQPublisher(rabbitClient)
	feedbackService := feedback.NewServiceWithPublisher(feedback.NewRepository(db), productService, feedbackPublisher)
	feedback.RegisterRoutes(r.Group("/api"), authService, feedbackService)
	flagflowCache := flagflow.NewRedisCache(redisClient, 5*time.Minute)
	flagflowPublisher := flagflow.NewRabbitMQPublisher(rabbitClient)
	flagflowService := flagflow.NewService(flagflow.NewRepository(db), productService, flagflowCache, flagflowPublisher)
	flagflow.RegisterRoutes(r.Group("/api"), authService, flagflowService)
	requirementService := requirement.NewService(requirement.NewRepository(db), productService)
	requirement.RegisterRoutes(r.Group("/api"), authService, requirementService)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("[%s] API server starting on %s (env=%s)", cfg.App.Name, addr, cfg.App.Env)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func main() {
	cfg, err := config.Load("internal/pkg/config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	StartHttpServer(cfg)
}
