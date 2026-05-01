// Package config provides configuration loading and parsing
// from files (config.yaml) and environment variables (.env).
package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

// Config holds all configuration sections for the application.
type Config struct {
	Logger  Logger  `mapstructure:"logger"`   // Logger configuration
	Server  Server  `mapstructure:"server"`   // Server HTTP settings
	Service Service `mapstructure:"service"`  // Service business logic settings
	Storage Storage `mapstructure:"database"` // Storage database settings
}

// Logger configures logging behavior.
type Logger struct {
	Debug  bool   `mapstructure:"debug_mode"`    // Debug enables debug logging
	LogDir string `mapstructure:"log_directory"` // LogDir is the directory for log files
}

// Server defines HTTP server parameters.
type Server struct {
	Port            string        `mapstructure:"port"`             // Port to listen on
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`     // ReadTimeout for HTTP requests
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`    // WriteTimeout for HTTP responses
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"` // MaxHeaderBytes limits request header size
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"` // ShutdownTimeout for graceful shutdown
}

// Service holds configurations for different service modules.
type Service struct {
	Auth Auth `mapstructure:"auth"` // Auth configuration
	Core Core `mapstructure:"core"` // Core business rules configuration
}

// Auth configures authentication and authorization.
type Auth struct {
	TokenSignedString string        // TokenSignedString is the secret for JWT signing
	TokenTTL          time.Duration `mapstructure:"token_ttl"`           // TokenTTL is the lifetime of JWT tokens
	MinLoginLength    int           `mapstructure:"min_login_length"`    // MinLoginLength for login username
	MaxLoginLength    int           `mapstructure:"max_login_length"`    // MaxLoginLength for login username
	MinPasswordLength int           `mapstructure:"min_password_length"` // MinPasswordLength for password
}

// Core configures core business logic constraints.
type Core struct {
	MinItemNameLength        int   `mapstructure:"min_item_name_length"`        // MinItemNameLength for item names
	MaxItemNameLength        int   `mapstructure:"max_item_name_length"`        // MaxItemNameLength for item names
	MaxItemDescriptionLength int   `mapstructure:"max_item_description_length"` // MaxItemDescriptionLength for item descriptions
	MinItemQuantity          int   `mapstructure:"min_item_quantity"`           // MinItemQuantity for item stock
	MaxItemQuantity          int   `mapstructure:"max_item_quantity"`           // MaxItemQuantity for item stock
	MaxItemPrice             int64 `mapstructure:"max_item_price"`              // MaxItemPrice in cents or smallest unit
}

// Storage configures database connection and migration settings.
type Storage struct {
	Dialect            string        `mapstructure:"goose_dialect"`              // Dialect for goose migrations
	MigrationsDir      string        `mapstructure:"goose_migrations_directory"` // MigrationsDir path for goose scripts
	Host               string        `mapstructure:"host"`                       // Host database server address
	Port               string        `mapstructure:"port"`                       // Port database server port
	Username           string        `mapstructure:"username"`                   // Username for database authentication
	Password           string        `mapstructure:"password"`                   // Password for database authentication
	DBName             string        `mapstructure:"dbname"`                     // DBName database name
	SSLMode            string        `mapstructure:"sslmode"`                    // SSLMode connection ssl mode
	MaxOpenConns       int           `mapstructure:"max_open_conns"`             // MaxOpenConns maximum open connections
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`             // MaxIdleConns maximum idle connections
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`          // ConnMaxLifetime per connection maximum lifetime
	QueryRetryStrategy RetryStrategy `mapstructure:"query_retry_strategy"`       // QueryRetryStrategy for retrying queries
	TxRetryStrategy    RetryStrategy `mapstructure:"tx_retry_strategy"`          // TxRetryStrategy for retrying transactions
}

// RetryStrategy defines retry behavior for operations.
type RetryStrategy struct {
	Attempts int           `mapstructure:"attempts"` // Attempts number of retry attempts
	Delay    time.Duration `mapstructure:"delay"`    // Delay initial delay between retries
	Backoff  float64       `mapstructure:"backoff"`  // Backoff multiplier for subsequent delays
}

// Load reads configuration files (config.yaml and .env) and environment variables,
// unmarshals them into a Config struct, and returns it.
// If the docker flag is not set, missing .env file causes an error.
func Load() (Config, error) {

	cfg := wbf.New()

	if err := cfg.LoadConfigFiles("./config.yaml"); err != nil {
		return Config{}, err
	}

	if err := cfg.LoadEnvFiles(".env"); err != nil && !cfg.GetBool("docker") {
		return Config{}, err
	}

	var conf Config

	if err := cfg.Unmarshal(&conf); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	loadEnvs(&conf)

	return conf, nil

}

// loadEnvs overrides specific configuration fields with environment variables.
// It reads DB_USER, DB_PASSWORD, and JWT_TOKEN_SIGNED_STRING from the environment.
func loadEnvs(conf *Config) {

	conf.Storage.Username = os.Getenv("DB_USER")
	conf.Storage.Password = os.Getenv("DB_PASSWORD")

	conf.Service.Auth.TokenSignedString = os.Getenv("JWT_TOKEN_SIGNED_STRING")

}
