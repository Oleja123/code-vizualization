package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerConfig      `yaml:"server"`
	OneCompilerConfig `yaml:"onecompiler"`
	LimitationsConfig `yaml:"limitations"`
	RedisConfig       `yaml:"redis"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type OneCompilerConfig struct {
	APIURL         string `yaml:"api_url"`
	APIKey         string `yaml:"api_key"`
	Enabled        bool   `yaml:"enabled"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type LimitationsConfig struct {
	MaxAllocatedElements int `yaml:"max_allocated_elements"`
	MaxSteps             int `yaml:"max_steps"`
}

type RedisConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	Expiration   time.Duration `yaml:"expiration"`
	PingAttempts int           `yaml:"ping_attempts"`
}

func Default() *Config {
	return &Config{
		ServerConfig: ServerConfig{
			Port: 8080,
		},
		OneCompilerConfig: OneCompilerConfig{
			APIURL:         "https://api.onecompiler.com/api/v1",
			Enabled:        false,
			TimeoutSeconds: 10,
		},
		LimitationsConfig: LimitationsConfig{
			MaxAllocatedElements: 100,
			MaxSteps:             1000,
		},
		RedisConfig: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           0,
			Expiration:   24 * time.Hour,
			PingAttempts: 3,
		},
	}
}

func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 10
	}

	if cfg.PingAttempts <= 0 {
		cfg.PingAttempts = 3
	}

	if cfg.MaxAllocatedElements <= 0 {
		cfg.MaxAllocatedElements = 100
	}

	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 1000
	}

	if cfg.Expiration <= 0 {
		cfg.Expiration = 24 * time.Hour
	}

	return cfg, nil
}

func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return Default()
	}

	return cfg
}
