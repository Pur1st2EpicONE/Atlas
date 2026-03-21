package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

type Config struct {
	Logger  Logger  `mapstructure:"logger"`
	Server  Server  `mapstructure:"server"`
	Service Service `mapstructure:"service"`
	Storage Storage `mapstructure:"database"`
}

type Logger struct {
	Debug  bool   `mapstructure:"debug_mode"`
	LogDir string `mapstructure:"log_directory"`
}

type Server struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Service struct {
	Auth Auth `mapstructure:"auth"`
	Core Core `mapstructure:"core"`
}

type Auth struct {
	TokenSignedString string
	TokenTTL          time.Duration `mapstructure:"token_ttl"`
	MinLoginLength    int           `mapstructure:"min_login_length"`
	MaxLoginLength    int           `mapstructure:"max_login_length"`
	MinPasswordLength int           `mapstructure:"min_password_length"`
}

type Core struct {
	MinItemNameLength        int   `mapstructure:"min_item_name_length"`
	MaxItemNameLength        int   `mapstructure:"max_item_name_length"`
	MaxItemDescriptionLength int   `mapstructure:"max_item_description_length"`
	MinItemQuantity          int   `mapstructure:"min_item_quantity"`
	MaxItemQuantity          int   `mapstructure:"max_item_quantity"`
	MaxItemPrice             int64 `mapstructure:"max_item_price"`
}

type Storage struct {
	Dialect            string        `mapstructure:"goose_dialect"`              // Goose migration dialect
	MigrationsDir      string        `mapstructure:"goose_migrations_directory"` // Directory for Goose migrations
	Host               string        `mapstructure:"host"`                       // DB host
	Port               string        `mapstructure:"port"`                       // DB port
	Username           string        `mapstructure:"username"`                   // DB username
	Password           string        `mapstructure:"password"`                   // DB password
	DBName             string        `mapstructure:"dbname"`                     // database name
	SSLMode            string        `mapstructure:"sslmode"`                    // SSL mode
	MaxOpenConns       int           `mapstructure:"max_open_conns"`             // maximum open connections
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`             // maximum idle connections
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`          // max lifetime per connection
	QueryRetryStrategy RetryStrategy `mapstructure:"query_retry_strategy"`       // query retry strategy
	TxRetryStrategy    RetryStrategy `mapstructure:"tx_retry_strategy"`          // transaction retry strategy
}

type RetryStrategy struct {
	Attempts int           `mapstructure:"attempts"`
	Delay    time.Duration `mapstructure:"delay"`
	Backoff  float64       `mapstructure:"backoff"`
}

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

func loadEnvs(conf *Config) {

	conf.Storage.Username = os.Getenv("DB_USER")
	conf.Storage.Password = os.Getenv("DB_PASSWORD")

	conf.Service.Auth.TokenSignedString = os.Getenv("JWT_TOKEN_SIGNED_STRING")

}
