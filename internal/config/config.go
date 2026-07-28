package config

import (
	"os"
	"time"
)

type Config struct {
	DBURL      string
	Port       string
	GinMode    string
	ConfigPath string
	TZ         *time.Location
}

func Load() *Config {
	cfg := &Config{
		DBURL:      getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/exam?sslmode=disable"),
		Port:       getEnv("PORT", "9000"),
		GinMode:    getEnv("GIN_MODE", "debug"),
		ConfigPath: getEnv("CONFIG_PATH", "./config/exam.json"),
	}

	tz := getEnv("TZ", "Asia/Shanghai")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	cfg.TZ = loc

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
