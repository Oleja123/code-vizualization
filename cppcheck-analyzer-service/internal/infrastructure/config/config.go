package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Cppcheck CppcheckConfig `yaml:"cppcheck"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type CppcheckConfig struct {
	Path           string `yaml:"path"`
	Std            string `yaml:"std"`
	Enable         string `yaml:"enable"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	Inconclusive   bool   `yaml:"inconclusive"`
	MaxIssues      int    `yaml:"max_issues"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	applyDefaults(&cfg)
	overrideWithEnv(&cfg)

	return &cfg, nil
}

func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		fallback := defaultConfig()
		overrideWithEnv(&fallback)
		return &fallback
	}

	return cfg
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8086
	}

	if cfg.Cppcheck.Path == "" {
		cfg.Cppcheck.Path = "cppcheck"
	}

	if cfg.Cppcheck.Std == "" {
		cfg.Cppcheck.Std = "c11"
	}

	if cfg.Cppcheck.Enable == "" {
		cfg.Cppcheck.Enable = "warning,style,performance,portability,information"
	}

	if cfg.Cppcheck.TimeoutSeconds <= 0 {
		cfg.Cppcheck.TimeoutSeconds = 15
	}

	if cfg.Cppcheck.MaxIssues <= 0 {
		cfg.Cppcheck.MaxIssues = 100
	}
}

func overrideWithEnv(cfg *Config) {
	if path := os.Getenv("CPPCHECK_PATH"); path != "" {
		cfg.Cppcheck.Path = path
	}

	if std := os.Getenv("CPPCHECK_STD"); std != "" {
		cfg.Cppcheck.Std = std
	}

	if enable := os.Getenv("CPPCHECK_ENABLE"); enable != "" {
		cfg.Cppcheck.Enable = enable
	}

	if timeout := os.Getenv("CPPCHECK_TIMEOUT_SECONDS"); timeout != "" {
		if v, err := strconv.Atoi(timeout); err == nil && v > 0 {
			cfg.Cppcheck.TimeoutSeconds = v
		}
	}

	if inconclusive := os.Getenv("CPPCHECK_INCONCLUSIVE"); inconclusive != "" {
		if v, err := strconv.ParseBool(inconclusive); err == nil {
			cfg.Cppcheck.Inconclusive = v
		}
	}

	if maxIssues := os.Getenv("CPPCHECK_MAX_ISSUES"); maxIssues != "" {
		if v, err := strconv.Atoi(maxIssues); err == nil && v > 0 {
			cfg.Cppcheck.MaxIssues = v
		}
	}
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{Port: 8086},
		Cppcheck: CppcheckConfig{
			Path:           "cppcheck",
			Std:            "c11",
			Enable:         "warning,style,performance,portability,information",
			TimeoutSeconds: 15,
			Inconclusive:   false,
			MaxIssues:      100,
		},
	}
}
