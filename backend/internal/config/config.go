package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	Port           string
	FrontendURL    string
	APIURL         string
	DatabaseURL    string
	RedisURL       string
	NatsURL        string
	JWTSecret      string
	EncryptionKey  string
	GitHubClientID string
	GitHubSecret   string
	DockerHost     string
	WorkspaceRoot  string
	DockerNetwork  string
	TraefikDomain  string
	Mode           string // "api" or "worker"
}

func Load() (*Config, error) {
	// Load .env file from project root
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	cfg := &Config{
		Env:            getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8082"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:3000"),
		APIURL:         getEnv("API_URL", "http://localhost:8082"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		NatsURL:        getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		EncryptionKey:  getEnv("ENCRYPTION_KEY", ""),
		GitHubClientID: getEnv("GITHUB_CLIENT_ID", ""),
		GitHubSecret:   getEnv("GITHUB_CLIENT_SECRET", ""),
		DockerHost:     getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		WorkspaceRoot:  getEnv("WORKSPACE_ROOT", "/tmp/harbor-workspaces"),
		DockerNetwork:  getEnv("DOCKER_NETWORK", "harbor"),
		TraefikDomain:  getEnv("TRAEFIK_DOMAIN", "localhost"),
		Mode:           getEnv("MODE", "api"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	return nil
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Env) == "development"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}