package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider manages configuration loading and validation
type Provider struct {
	config     *Config
	configPath string
}

// NewProvider creates a new configuration provider
func NewProvider(configPath string) *Provider {
	return &Provider{
		configPath: configPath,
	}
}

// Load loads configuration from file
func (p *Provider) Load() error {
	// Check if file exists
	if _, err := os.Stat(p.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", p.configPath)
	}

	// Read file
	data, err := os.ReadFile(p.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := p.validate(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	p.config = config
	return nil
}

// Get returns the loaded configuration
func (p *Provider) Get() *Config {
	return p.config
}

// validate performs configuration validation
func (p *Provider) validate(config *Config) error {
	// Validate Database settings
	if config.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database.maxOpenConns must be positive")
	}
	if config.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("database.maxIdleConns must be positive")
	}
	if config.Database.MaxLifetime <= 0 {
		return fmt.Errorf("database.maxLifetime must be positive")
	}

	// Validate HTTP settings
	if config.HTTP.Port <= 0 || config.HTTP.Port > 65535 {
		return fmt.Errorf("http.port must be between 1 and 65535")
	}
	if config.HTTP.ReadTimeout <= 0 {
		return fmt.Errorf("http.readTimeout must be positive")
	}
	if config.HTTP.WriteTimeout <= 0 {
		return fmt.Errorf("http.writeTimeout must be positive")
	}

	// Validate Logger settings
	level := strings.ToLower(config.Logger.Level)
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[level] {
		return fmt.Errorf("invalid logger.level: %s", config.Logger.Level)
	}

	// Validate Metrics settings
	if config.Metrics.Enabled {
		if config.Metrics.Endpoint == "" {
			return fmt.Errorf("metrics.endpoint is required when metrics are enabled")
		}
		if config.Metrics.Interval <= 0 {
			return fmt.Errorf("metrics.interval must be positive")
		}
	}

	return nil
}

// GetConfigPath returns the absolute path for a config file
func (p *Provider) GetConfigPath(env string) string {
	if env == "" {
		env = "development"
	}
	filename := fmt.Sprintf("config.%s.json", env)
	return filepath.Join(filepath.Dir(p.configPath), filename)
}
