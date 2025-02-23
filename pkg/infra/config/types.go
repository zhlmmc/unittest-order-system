package config

import (
	"time"
)

// Config represents the configuration settings
type Config struct {
	// Database settings
	Database struct {
		Host         string        `json:"host"`
		Port         int           `json:"port"`
		User         string        `json:"user"`
		Password     string        `json:"password"`
		Database     string        `json:"database"`
		MaxOpenConns int           `json:"maxOpenConns"`
		MaxIdleConns int           `json:"maxIdleConns"`
		MaxLifetime  time.Duration `json:"maxLifetime"`
	} `json:"database"`

	// HTTP settings
	HTTP struct {
		Port            int           `json:"port"`
		ReadTimeout     time.Duration `json:"readTimeout"`
		WriteTimeout    time.Duration `json:"writeTimeout"`
		MaxHeaderBytes  int           `json:"maxHeaderBytes"`
		MaxRequestSize  int64         `json:"maxRequestSize"`
		RequestTimeout  time.Duration `json:"requestTimeout"`
		ShutdownTimeout time.Duration `json:"shutdownTimeout"`
	} `json:"http"`

	// Logger settings
	Logger struct {
		Level      string `json:"level"`
		Format     string `json:"format"`
		Output     string `json:"output"`
		TimeFormat string `json:"timeFormat"`
	} `json:"logger"`

	// Metrics settings
	Metrics struct {
		Enabled     bool          `json:"enabled"`
		Endpoint    string        `json:"endpoint"`
		PushGateway string        `json:"pushGateway"`
		Interval    time.Duration `json:"interval"`
	} `json:"metrics"`
}
