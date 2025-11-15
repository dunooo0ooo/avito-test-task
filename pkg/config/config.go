package config

import (
	"os"
	"strconv"
)

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	SSLMode       string
	MaxConns      int32
	MinConns      int32
	MigrationsDir string
}

type LoggerConfig struct {
	Level string
}

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Logger   LoggerConfig
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func getenvInt32(key string, def int32) int32 {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return int32(n)
		}
	}
	return def
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: getenv("HTTP_ADDR", ":8080"),
		},
		Postgres: PostgresConfig{
			Host:          getenv("POSTGRES_HOST", "localhost"),
			Port:          getenvInt("POSTGRES_PORT", 5432),
			User:          getenv("POSTGRES_USER", "postgres"),
			Password:      getenv("POSTGRES_PASSWORD", "postgres"),
			DBName:        getenv("POSTGRES_DB", "app"),
			SSLMode:       getenv("POSTGRES_SSLMODE", "disable"),
			MaxConns:      getenvInt32("POSTGRES_MAX_CONNS", 10),
			MinConns:      getenvInt32("POSTGRES_MIN_CONNS", 2),
			MigrationsDir: getenv("MIGRATIONS_DIR", "./migrations"),
		},
		Logger: LoggerConfig{
			Level: getenv("LOG_LEVEL", "info"),
		},
	}
}

func (p PostgresConfig) DSN() string {
	return "postgres://" + p.User + ":" + p.Password +
		"@" + p.Host + ":" + strconv.Itoa(p.Port) +
		"/" + p.DBName + "?sslmode=" + p.SSLMode
}
