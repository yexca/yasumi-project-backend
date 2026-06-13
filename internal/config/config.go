package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv    string
	HTTP      HTTPConfig
	Log       LogConfig
	Postgres  PostgresConfig
	PowerSync PowerSyncConfig
}

type HTTPConfig struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func (c HTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

type LogConfig struct {
	Level  string
	Format string
}

type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

func (c PostgresConfig) DSN() string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.Database,
	}
	query := dsn.Query()
	query.Set("sslmode", c.SSLMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

type PowerSyncConfig struct {
	URL string
}

func Load() (Config, error) {
	httpPort, httpPortErr := getInt("YASUMI_HTTP_PORT", 8080)
	readHeaderTimeout, readHeaderTimeoutErr := getDuration("YASUMI_HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	shutdownTimeout, shutdownTimeoutErr := getDuration("YASUMI_HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	postgresPort, postgresPortErr := getInt("YASUMI_POSTGRES_PORT", 5432)

	var parseProblems []string
	for _, err := range []error{httpPortErr, readHeaderTimeoutErr, shutdownTimeoutErr, postgresPortErr} {
		if err != nil {
			parseProblems = append(parseProblems, err.Error())
		}
	}
	if len(parseProblems) > 0 {
		return Config{}, errors.New(strings.Join(parseProblems, "; "))
	}

	cfg := Config{
		AppEnv: getString("YASUMI_APP_ENV", "local"),
		HTTP: HTTPConfig{
			Host:              getString("YASUMI_HTTP_HOST", "0.0.0.0"),
			Port:              httpPort,
			ReadHeaderTimeout: readHeaderTimeout,
			ShutdownTimeout:   shutdownTimeout,
		},
		Log: LogConfig{
			Level:  getString("YASUMI_LOG_LEVEL", "info"),
			Format: getString("YASUMI_LOG_FORMAT", "json"),
		},
		Postgres: PostgresConfig{
			Host:     getString("YASUMI_POSTGRES_HOST", "localhost"),
			Port:     postgresPort,
			Database: getString("YASUMI_POSTGRES_DB", "yasumi"),
			User:     getString("YASUMI_POSTGRES_USER", "yasumi"),
			Password: getString("YASUMI_POSTGRES_PASSWORD", "yasumi"),
			SSLMode:  getString("YASUMI_POSTGRES_SSLMODE", "disable"),
		},
		PowerSync: PowerSyncConfig{
			URL: getString("YASUMI_POWERSYNC_URL", "http://localhost:8081"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	var problems []string

	if c.AppEnv == "" {
		problems = append(problems, "YASUMI_APP_ENV must not be empty")
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		problems = append(problems, "YASUMI_HTTP_PORT must be between 1 and 65535")
	}
	if c.HTTP.ReadHeaderTimeout <= 0 {
		problems = append(problems, "YASUMI_HTTP_READ_HEADER_TIMEOUT must be positive")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		problems = append(problems, "YASUMI_HTTP_SHUTDOWN_TIMEOUT must be positive")
	}
	if !isOneOf(c.Log.Level, "debug", "info", "warn", "error") {
		problems = append(problems, "YASUMI_LOG_LEVEL must be one of debug, info, warn, error")
	}
	if !isOneOf(c.Log.Format, "json", "text") {
		problems = append(problems, "YASUMI_LOG_FORMAT must be one of json, text")
	}
	if c.Postgres.Host == "" {
		problems = append(problems, "YASUMI_POSTGRES_HOST must not be empty")
	}
	if c.Postgres.Port < 1 || c.Postgres.Port > 65535 {
		problems = append(problems, "YASUMI_POSTGRES_PORT must be between 1 and 65535")
	}
	if c.Postgres.Database == "" {
		problems = append(problems, "YASUMI_POSTGRES_DB must not be empty")
	}
	if c.Postgres.User == "" {
		problems = append(problems, "YASUMI_POSTGRES_USER must not be empty")
	}
	if c.Postgres.SSLMode == "" {
		problems = append(problems, "YASUMI_POSTGRES_SSLMODE must not be empty")
	}
	if c.PowerSync.URL == "" {
		problems = append(problems, "YASUMI_POWERSYNC_URL must not be empty")
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}

	return nil
}

func getString(name, fallback string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(value)
}

func getInt(name string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return value, nil
}

func getDuration(name string, fallback time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a Go duration such as 5s or 1m", name)
	}
	return value, nil
}

func isOneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}
	return cfg
}
