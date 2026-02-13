package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config представляет конфигурацию приложения
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	OneCompiler OneCompilerConfig `yaml:"onecompiler"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port int `yaml:"port"`
}

// OneCompilerConfig конфигурация OneCompiler API
type OneCompilerConfig struct {
	APIURL         string `yaml:"api_url"`
	APIKey         string `yaml:"api_key"`
	Enabled        bool   `yaml:"enabled"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Применяем значения по умолчанию
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if cfg.OneCompiler.TimeoutSeconds == 0 {
		cfg.OneCompiler.TimeoutSeconds = 10
	}

	// Позволяем переопределять API ключ через переменную окружения
	if apiKey := os.Getenv("ONECOMPILER_API_KEY"); apiKey != "" {
		cfg.OneCompiler.APIKey = apiKey
	}

	return &cfg, nil
}

// LoadConfigOrDefault загружает конфиг или возвращает конфиг по умолчанию
func LoadConfigOrDefault(path string) *Config {
	cfg, err := LoadConfig(path)
	if err != nil {
		return &Config{
			Server: ServerConfig{Port: 8080},
			OneCompiler: OneCompilerConfig{
				APIURL:         "https://api.onecompiler.com/api/v1",
				Enabled:        true,
				TimeoutSeconds: 10,
			},
		}
	}
	return cfg
}
