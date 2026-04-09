// Package config defines all configuration structures for the Proteus
// image processing service and provides the Load function that assembles
// the complete configuration from YAML files, .env files and environment
// variables (using the wb-go/config library).
package config

import (
	"fmt"
	"os"
	"time"

	wbf "github.com/wb-go/wbf/config"
)

type Config struct {
	Logger     Logger     `mapstructure:"logger"`     // Logger holds logging-related settings for the whole application.
	Server     Server     `mapstructure:"server"`     // Server contains HTTP server configuration parameters (port, timeouts, limits).
	Service    Service    `mapstructure:"service"`    // Service holds business-logic specific settings such as the cleaner.
	Consumer   Consumer   `mapstructure:"consumer"`   // Consumer defines Kafka consumer settings (brokers, topic, group, retry policy).
	Repository Repository `mapstructure:"repository"` // Repository groups meta (PostgreSQL) and image (MinIO) storage configurations.
}

type Logger struct {
	Debug  bool   `mapstructure:"debug_mode"`    // Debug enables debug-level logging when true.
	LogDir string `mapstructure:"log_directory"` // LogDir is the directory where log files are stored (ignored when logging to stdout).
}

type Server struct {
	Port            string        `mapstructure:"port"`             // Port is the TCP port the HTTP server listens on.
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`     // ReadTimeout is the maximum duration for reading the full request.
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`    // WriteTimeout is the maximum duration before timing out writes.
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"` // MaxHeaderBytes limits the size of request headers.
	MaxFileSize     int64         `mapstructure:"max_file_size"`    // MaxFileSize is the maximum allowed size of a single uploaded file.
	MaxRequestSize  int64         `mapstructure:"max_request_size"` // MaxRequestSize limits the total size of the request body.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"` // ShutdownTimeout is the grace period given to the server during shutdown.
}

type Consumer struct {
	Brokers            []string      `mapstructure:"brokers"`              // Brokers is the list of Kafka broker addresses.
	Topic              string        `mapstructure:"topic"`                // Topic is the Kafka topic from which image processing tasks are consumed.
	GroupID            string        `mapstructure:"group_id"`             // GroupID is the Kafka consumer group identifier.
	FetchRetryStrategy RetryStrategy `mapstructure:"fetch_retry_strategy"` // FetchRetryStrategy configures retries when fetching messages from Kafka.
}

type Repository struct {
	MetaStorage  MetaStorage  `mapstructure:"meta_storage"`  // MetaStorage holds PostgreSQL connection and migration settings.
	ImageStorage ImageStorage `mapstructure:"image_storage"` // ImageStorage holds MinIO object storage configuration.
}

type MetaStorage struct {
	Dialect            string        `mapstructure:"goose_dialect"`              // Dialect is the database dialect passed to Goose for migrations.
	MigrationsDir      string        `mapstructure:"goose_migrations_directory"` // MigrationsDir is the directory containing Goose migration files.
	Host               string        `mapstructure:"host"`                       // Host is the PostgreSQL server hostname or IP.
	Port               string        `mapstructure:"port"`                       // Port is the PostgreSQL server port.
	Username           string        `mapstructure:"username"`                   // Username is the PostgreSQL user (overridden from env).
	Password           string        `mapstructure:"password"`                   // Password is the PostgreSQL password (overridden from env).
	DBName             string        `mapstructure:"dbname"`                     // DBName is the name of the PostgreSQL database.
	SSLMode            string        `mapstructure:"sslmode"`                    // SSLMode controls SSL behavior for the PostgreSQL connection.
	MaxOpenConns       int           `mapstructure:"max_open_conns"`             // MaxOpenConns is the maximum number of open connections to the DB.
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`             // MaxIdleConns is the maximum number of idle connections.
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`          // ConnMaxLifetime is the maximum lifetime of a connection.
	PendingTimeout     time.Duration `yaml:"pending_timeout"`                    // PendingTimeout is the timeout for pending operations (used by the service layer).
	QueryRetryStrategy RetryStrategy `mapstructure:"query_retry_strategy"`       // QueryRetryStrategy configures retries for database queries.
}

type Service struct {
	Cleaner         bool          `mapstructure:"cleaner"`          // Cleaner enables the background cleaner goroutine when true.
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"` // CleanupInterval defines how often the cleaner runs.
}

type ImageStorage struct {
	MinIOEndpoint  string `mapstructure:"minio_endpoint"` // MinIOEndpoint is the MinIO server address.
	MinIOAccessKey string // MinIOAccessKey is the MinIO access key (loaded only from env).
	MinIOSecretKey string // MinIOSecretKey is the MinIO secret key (loaded only from env).
	MinIOBucket    string `mapstructure:"minio_bucket"`  // MinIOBucket is the name of the bucket used for images.
	MinIOUseSSL    bool   `mapstructure:"minio_use_ssl"` // MinIOUseSSL enables HTTPS when connecting to MinIO.
	MinIORegion    string `mapstructure:"minio_region"`  // MinIORegion is the region passed to the MinIO client.
}

type RetryStrategy struct {
	Attempts int           `mapstructure:"attempts"` // Attempts is the maximum number of retry attempts.
	Delay    time.Duration `mapstructure:"delay"`    // Delay is the initial delay between retries.
	Backoff  float64       `mapstructure:"backoff"`  // Backoff is the multiplier for exponential backoff.
}

// Load reads the application configuration from ./config.yaml (and optionally
// .env) using the wb-go/config library, unmarshals it into Config and then
// overrides sensitive credentials from environment variables. It returns the
// fully populated configuration or an error.
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

// loadEnvs overrides database credentials and MinIO keys in the already
// unmarshaled config with values taken directly from environment variables.
// It is called after unmarshaling so that secrets never appear in config files.
func loadEnvs(conf *Config) {

	conf.Repository.MetaStorage.Username = os.Getenv("DB_USER")
	conf.Repository.MetaStorage.Password = os.Getenv("DB_PASSWORD")

	conf.Repository.ImageStorage.MinIOAccessKey = os.Getenv("MINIO_ROOT_USER")
	conf.Repository.ImageStorage.MinIOSecretKey = os.Getenv("MINIO_ROOT_PASSWORD")

}
