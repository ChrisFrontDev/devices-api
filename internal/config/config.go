package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Server   ServerConfig   `yaml:"server"`
		Database DatabaseConfig `yaml:"database"`
	}

	ServerConfig struct {
		HTTPPort int `yaml:"http_port" env:"SERVER_HTTP_PORT" env-default:"8080"`
		GRPCPort int `yaml:"grpc_port" env:"SERVER_GRPC_PORT" env-default:"9090"`
	}

	DatabaseConfig struct {
		URL string `yaml:"url" env:"DATABASE_URL" env-required:"true"`
	}
)

// LoadConfig loads configuration from environment variables.
// Note: We deliberately do NOT automatically load .env files here to strictly follow
// 12-factor app principles. In Docker, env vars are injected by the runtime.
// For local development without Docker, export variables in your shell or use a tool like direnv.
func LoadConfig() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return &cfg, nil
}
