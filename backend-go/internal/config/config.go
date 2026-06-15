package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppName string
	Env     string
	Debug   bool
	Port    int
	SecretKey string
	AccessTokenExpireMinutes int

	DBDriver string // "sqlite" or "mysql"
	DBPath   string // for sqlite
	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string

	MaxUploadSize int64
}

func Load() *Config {
	cfg := &Config{
		AppName:   getEnv("APP_NAME", "NoteWeb API"),
		Env:       getEnv("ENV", "development"),
		Debug:     getEnv("DEBUG", "true") == "true",
		Port:      getEnvInt("PORT", 8000),
		SecretKey: getEnv("SECRET_KEY", "change-me-in-production"),
		AccessTokenExpireMinutes: getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES", 60*24),

		DBDriver: getEnv("DB_DRIVER", "sqlite"),
		DBPath:   getEnv("DB_PATH", "./noteweb.db"),

		MySQLHost:     getEnv("MYSQL_HOST", "localhost"),
		MySQLPort:     getEnvInt("MYSQL_PORT", 3306),
		MySQLUser:     getEnv("MYSQL_USER", "root"),
		MySQLPassword: getEnv("MYSQL_PASSWORD", "123456"),
		MySQLDatabase: getEnv("MYSQL_DATABASE", "noteweb"),

		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 50*1024*1024),
	}
	return cfg
}

func (c *Config) DSN() string {
	if c.DBDriver == "sqlite" {
		return c.DBPath
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQLUser, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}
