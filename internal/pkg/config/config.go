package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type RabbitMQConfig struct {
	URL string `yaml:"url"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := loadFromFile(configPath, cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateJWT(cfg *Config) error {
	if cfg.App.Env == "development" {
		return nil
	}

	if cfg.JWT.Secret == "change-me-in-production" {
		return fmt.Errorf("jwt.secret is still the default placeholder — set PULSEROAD_JWT_SECRET to a real secret in production")
	}

	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("jwt.secret is too short (%d chars), minimum 32 chars required for HS256 — set PULSEROAD_JWT_SECRET", len(cfg.JWT.Secret))
	}

	return nil
}

func loadFromFile(path string, cfg *Config) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	return nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PULSEROAD_APP_NAME"); v != "" {
		cfg.App.Name = v
	}
	if v := os.Getenv("PULSEROAD_APP_ENV"); v != "" {
		cfg.App.Env = v
	}
	if v := os.Getenv("PULSEROAD_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("PULSEROAD_MYSQL_DSN"); v != "" {
		cfg.MySQL.DSN = v
	}
	if v := os.Getenv("PULSEROAD_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("PULSEROAD_RABBITMQ_URL"); v != "" {
		cfg.RabbitMQ.URL = v
	}
	if v := os.Getenv("PULSEROAD_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
}

func validate(cfg *Config) error {
	var missing []string

	if cfg.App.Name == "" {
		missing = append(missing, "app.name")
	}
	if cfg.App.Env == "" {
		missing = append(missing, "app.env")
	}
	if cfg.Server.Port == 0 {
		missing = append(missing, "server.port")
	}
	if cfg.MySQL.DSN == "" {
		missing = append(missing, "mysql.dsn")
	}
	if cfg.Redis.Addr == "" {
		missing = append(missing, "redis.addr")
	}
	if cfg.RabbitMQ.URL == "" {
		missing = append(missing, "rabbitmq.url")
	}
	if cfg.JWT.Secret == "" {
		missing = append(missing, "jwt.secret")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}

	if err := validateJWT(cfg); err != nil {
		return err
	}

	return nil
}
