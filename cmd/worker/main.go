package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"pulseroad/internal/feedback"
	"pulseroad/internal/flagflow"
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

	rabbitClient, err := rabbitmq.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	defer func() {
		if err := rabbitClient.Close(); err != nil {
			log.Printf("failed to close rabbitmq connection: %v", err)
		}
	}()

	log.Printf("[%s] Worker process started successfully (env=%s)", cfg.App.Name, cfg.App.Env)
	log.Printf("RabbitMQ consumer registered for %s", feedback.QueueFeedbackCreated)
	log.Printf("RabbitMQ consumer registered for %s", flagflow.QueueFlagChanged)
	ctx := context.Background()
	go func() {
		if err := rabbitClient.ConsumeJSON(ctx, feedback.QueueFeedbackCreated, handleFeedbackCreatedMessage); err != nil {
			log.Fatalf("feedback consumer stopped: %v", err)
		}
	}()
	if err := rabbitClient.ConsumeJSON(ctx, flagflow.QueueFlagChanged, handleFlagChangedMessage); err != nil {
		log.Fatalf("flagflow consumer stopped: %v", err)
	}
}

func handleFeedbackCreatedMessage(_ context.Context, body []byte) error {
	var event feedback.FeedbackCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("decode feedback created event: %w", err)
	}

	log.Printf(
		"feedback.created feedback_id=%d product_id=%d team_id=%d created_by=%d title=%q status=%s",
		event.FeedbackID,
		event.ProductID,
		event.TeamID,
		event.CreatedBy,
		event.Title,
		event.Status,
	)
	return nil
}

func handleFlagChangedMessage(_ context.Context, body []byte) error {
	var event flagflow.FlagChangedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("decode flag changed event: %w", err)
	}

	log.Printf(
		"flag.changed flag_id=%d product_id=%d team_id=%d key=%q environment=%s enabled=%t rollout=%d action=%s changed_by=%d",
		event.FlagID,
		event.ProductID,
		event.TeamID,
		event.Key,
		event.Environment,
		event.Enabled,
		event.RolloutPercentage,
		event.Action,
		event.ChangedBy,
	)
	return nil
}

func main() {
	// 加载配置
	cfg, err := config.Load("internal/pkg/config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	StartWorker(cfg)
}
