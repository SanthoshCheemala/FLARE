package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	PSI      PSIConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Driver   string // sqlite or postgres
	DSN      string // Database connection string
	MaxConns int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type PSIConfig struct {
	TreeDBPath    string
	MaxRAMGB      float64
	MaxWorkers    int
	MaxScreenings int
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite3"),
			DSN:      getEnv("DB_DSN", "./data/flare.db"),
			MaxConns: getIntEnv("DB_MAX_CONNS", 25),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "change-this-secret"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "change-this-refresh-secret"),
			AccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "flare-api"),
		},
		PSI: PSIConfig{
			TreeDBPath:    getEnv("PSI_TREE_PATH", "./data/trees"),
			MaxRAMGB:      getFloatEnv("PSI_MAX_RAM_GB", 16.0),
			MaxWorkers:    getIntEnv("PSI_MAX_WORKERS", 0), // 0 = auto
			MaxScreenings: getIntEnv("PSI_MAX_CONCURRENT_SCREENINGS", 2),
		},
		Redis: RedisConfig{
			Enabled:  getBoolEnv("REDIS_ENABLED", false),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func (c *Config) DatabaseDSN() string {
	if c.Database.Driver == "sqlite3" {
		if !strings.Contains(c.Database.DSN, "?") {
			return c.Database.DSN + "?_parseTime=true"
		} else if !strings.Contains(c.Database.DSN, "_parseTime") {
			return c.Database.DSN + "&_parseTime=true"
		}
	}
	return c.Database.DSN
}

func (c *Config) DatabaseDriver() string {
	return c.Database.Driver
}
