package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

type Config struct {
	Logger     Logger     `mapstructure:"logger"`
	Server     Server     `mapstructure:"server"`
	Consumer   Consumer   `mapstructure:"consumer"`
	Producer   Producer   `mapstructure:"producer"`
	Repository Repository `mapstructure:"repository"`
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

type Consumer struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	GroupID string   `mapstructure:"group_id"`
}

type Producer struct {
}

type Repository struct {
	MetaStorage  MetaStorage  `mapstructure:"meta_storage"`
	ImageStorage ImageStorage `mapstructure:"image_storage"`
}

type MetaStorage struct {
	Dialect            string        `mapstructure:"goose_dialect"`              // Goose migration dialect
	MigrationsDir      string        `mapstructure:"goose_migrations_directory"` // Directory for Goose migrations
	Host               string        `mapstructure:"host"`
	Port               string        `mapstructure:"port"`
	Username           string        `mapstructure:"username"`
	Password           string        `mapstructure:"password"`
	DBName             string        `mapstructure:"dbname"`
	SSLMode            string        `mapstructure:"sslmode"`
	MaxOpenConns       int           `mapstructure:"max_open_conns"`
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
	QueryRetryStrategy RetryStrategy `mapstructure:"query_retry_strategy"`
}

type ImageStorage struct {
	MinIOEndpoint  string `mapstructure:"minio_endpoint"`
	MinIOAccessKey string `mapstructure:"minio_access_key"`
	MinIOSecretKey string `mapstructure:"minio_secret_key"`
	MinIOBucket    string `mapstructure:"minio_bucket"`
	MinIOUseSSL    bool   `mapstructure:"minio_use_ssl"`
	MinIORegion    string `mapstructure:"minio_region"`
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

	conf.Repository.MetaStorage.Username = os.Getenv("DB_USER")
	conf.Repository.MetaStorage.Password = os.Getenv("DB_PASSWORD")

}
