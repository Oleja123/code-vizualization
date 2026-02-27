package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerConfig      `yaml:"server"`
	OneCompilerConfig `yaml:"onecompiler"`
	LimitationsConfig `yaml:"limitations"`
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

	return cfg, nil
}

func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return Default()
	}

	return cfg
}
