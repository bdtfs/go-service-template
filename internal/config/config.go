package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service    ServiceConfig    `yaml:"service"`
	Server     ServerConfig     `yaml:"server"`
	Log        LogConfig        `yaml:"log"`
	Metrics    MetricsConfig    `yaml:"metrics"`
	Components ComponentsConfig `yaml:"components"`
	App        AppConfig        `yaml:"app"`
}

// AppConfig holds application-domain tunables (downstream endpoints, per-use-case
// limits). Infrastructure lives in the sections above; put business knobs here so
// the di.Container can read them when building use cases.
type AppConfig struct {
	Notifier        NotifierConfig        `yaml:"notifier"`
	CreateOperation CreateOperationConfig `yaml:"create_operation"`
}

type NotifierConfig struct {
	BaseURL string   `yaml:"base_url"`
	Timeout Duration `yaml:"timeout"`
}

type CreateOperationConfig struct {
	MaxReattempt int64 `yaml:"max_reattempt"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type ServerConfig struct {
	Port         string   `yaml:"port"`
	ReadTimeout  Duration `yaml:"read_timeout"`
	WriteTimeout Duration `yaml:"write_timeout"`
	IdleTimeout  Duration `yaml:"idle_timeout"`
}

type LogConfig struct {
	Level     LogLevel `yaml:"level"`
	Format    string   `yaml:"format"`
	Dest      string   `yaml:"dest"`
	AddSource bool     `yaml:"add_source"`
}

func (c LogConfig) Writer() io.Writer {
	switch c.Dest {
	case "stderr":
		return os.Stderr
	default:
		return os.Stdout
	}
}

func (c LogConfig) SlogLevel() slog.Level {
	return slog.Level(c.Level)
}

type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Address   string `yaml:"address"`
	Namespace string `yaml:"namespace"`
	Subsystem string `yaml:"subsystem"`
}

type ComponentsConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
}

type PostgresConfig struct {
	Enabled  bool   `yaml:"enabled"`
	DSN      string `yaml:"dsn"`
	MaxConns int32  `yaml:"max_conns"`
	MinConns int32  `yaml:"min_conns"`
}

// Duration wraps time.Duration for YAML unmarshalling from strings like "5s".
type Duration time.Duration

func (d Duration) Std() time.Duration {
	return time.Duration(d)
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(dur)
	return nil
}

// LogLevel wraps slog.Level for YAML unmarshalling from strings like "info".
type LogLevel slog.Level

func (l *LogLevel) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "debug":
		*l = LogLevel(slog.LevelDebug)
	case "info":
		*l = LogLevel(slog.LevelInfo)
	case "warn", "warning":
		*l = LogLevel(slog.LevelWarn)
	case "error":
		*l = LogLevel(slog.LevelError)
	default:
		return fmt.Errorf("unknown log level: %s", s)
	}
	return nil
}

// Load reads a YAML config file and interpolates environment variables.
// Supports ${VAR} and ${VAR:default} syntax.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	expanded := os.Expand(string(data), envLookup)

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Service.Name == "" {
		c.Service.Name = "service"
	}
	if c.Service.Type == "" {
		c.Service.Type = "api"
	}
	if c.Server.Port == "" {
		c.Server.Port = ":8080"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = Duration(5 * time.Second)
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = Duration(10 * time.Second)
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = Duration(120 * time.Second)
	}
	if c.Metrics.Address == "" {
		c.Metrics.Address = ":8081"
	}
	if c.Metrics.Namespace == "" {
		c.Metrics.Namespace = c.Service.Name
	}
	if c.Metrics.Subsystem == "" {
		c.Metrics.Subsystem = "app"
	}
	if c.Log.Dest == "" {
		c.Log.Dest = "stdout"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
	if c.App.Notifier.BaseURL == "" {
		c.App.Notifier.BaseURL = "http://localhost:9000"
	}
	if c.App.Notifier.Timeout == 0 {
		c.App.Notifier.Timeout = Duration(5 * time.Second)
	}
	if c.App.CreateOperation.MaxReattempt == 0 {
		c.App.CreateOperation.MaxReattempt = 5
	}
}

func (c *Config) IsProd() bool {
	return strings.EqualFold(os.Getenv("APP_ENVIRONMENT"), "prod")
}

func (c *Config) IsStaging() bool {
	return strings.EqualFold(os.Getenv("APP_ENVIRONMENT"), "staging")
}

func envLookup(key string) string {
	parts := strings.SplitN(key, ":", 2)
	name := parts[0]

	if val, ok := os.LookupEnv(name); ok {
		return val
	}
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// ParseBool is a helper for parsing boolean env vars / config values.
func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
