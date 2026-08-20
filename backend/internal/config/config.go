package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment / file.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Admin     AdminConfig     `mapstructure:"admin"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	AllowOrigins    []string      `mapstructure:"allow_origins"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"issuer"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RateLimitConfig struct {
	LoginPerMinuteIP   int `mapstructure:"login_per_minute_ip"`
	LoginPerMinuteUser int `mapstructure:"login_per_minute_user"`
	GlobalPerMinute     int `mapstructure:"global_per_minute"`
}

// Load reads configuration from environment variables (prefixed APP_) and
// optionally a config file. Environment variables take precedence.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.shutdown_timeout", "30s")
	v.SetDefault("server.allow_origins", "http://localhost:8080,http://localhost:5173")

	v.SetDefault("mysql.host", "mysql")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.user", "league_user")
	v.SetDefault("mysql.password", "league_pass")
	v.SetDefault("mysql.database", "league_db")
	v.SetDefault("mysql.max_open_conns", 50)
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.conn_max_lifetime", "5m")

	v.SetDefault("redis.host", "redis")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("jwt.secret", "change-me-in-production-please-use-long-random-string")
	v.SetDefault("jwt.access_token_ttl", "15m")
	v.SetDefault("jwt.refresh_token_ttl", "168h")
	v.SetDefault("jwt.issuer", "go-sports-league")

	v.SetDefault("admin.username", "admin")
	v.SetDefault("admin.password", "Admin123!")
	v.SetDefault("admin.email", "admin@league.local")

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	v.SetDefault("rate_limit.login_per_minute_ip", 5)
	v.SetDefault("rate_limit.login_per_minute_user", 10)
	v.SetDefault("rate_limit.global_per_minute", 600)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	// fail-safe: never allow zero rate limits (would block everything)
	if cfg.RateLimit.LoginPerMinuteIP <= 0 {
		cfg.RateLimit.LoginPerMinuteIP = 5
	}
	if cfg.RateLimit.LoginPerMinuteUser <= 0 {
		cfg.RateLimit.LoginPerMinuteUser = 10
	}
	if cfg.RateLimit.GlobalPerMinute <= 0 {
		cfg.RateLimit.GlobalPerMinute = 600
	}
	return cfg, nil
}
