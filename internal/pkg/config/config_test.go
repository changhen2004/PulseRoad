package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKeepsRedisAndRabbitMQConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
app:
  name: "pulseroad"
  env: "development"
server:
  port: 8080
mysql:
  dsn: "user:password@tcp(127.0.0.1:3306)/pulseroad?charset=utf8mb4&parseTime=True&loc=Local"
redis:
  addr: "127.0.0.1:6379"
rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
jwt:
  secret: "change-me-in-production"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.RabbitMQ.URL != "amqp://guest:guest@127.0.0.1:5672/" {
		t.Fatalf("unexpected rabbitmq url: %q", cfg.RabbitMQ.URL)
	}
	if cfg.Redis.Addr != "127.0.0.1:6379" {
		t.Fatalf("unexpected redis addr: %q", cfg.Redis.Addr)
	}
}

func TestRabbitMQURLEnvOverride(t *testing.T) {
	t.Setenv("PULSEROAD_RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
app:
  name: "pulseroad"
  env: "development"
server:
  port: 8080
mysql:
  dsn: "user:password@tcp(127.0.0.1:3306)/pulseroad?charset=utf8mb4&parseTime=True&loc=Local"
redis:
  addr: "127.0.0.1:6379"
rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
jwt:
  secret: "change-me-in-production"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.RabbitMQ.URL != "amqp://guest:guest@rabbitmq:5672/" {
		t.Fatalf("expected env rabbitmq url override, got %q", cfg.RabbitMQ.URL)
	}
}
