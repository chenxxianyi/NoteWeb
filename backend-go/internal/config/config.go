package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName string
	Env     string
	Debug   bool
	Port    int
	SecretKey string
	AccessTokenExpireMinutes int

	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string

	MaxUploadSize int64

	// AI Provider配置
	DeepSeekAPIKey  string // 默认DeepSeek API Key(可选,用户可覆盖)
	DeepSeekBaseURL string // DeepSeek API地址
}

func Load() *Config {
	loadDotEnv(".env")

	cfg := &Config{
		AppName:   getEnv("APP_NAME", "NoteWeb API"),
		Env:       getEnv("ENV", "development"),
		Debug:     getEnv("DEBUG", "true") == "true",
		Port:      getEnvInt("PORT", 8020),
		SecretKey: getEnv("SECRET_KEY", "change-me-in-production"),
		AccessTokenExpireMinutes: getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES", 60*24),

		MySQLHost:     getEnv("MYSQL_HOST", "localhost"),
		MySQLPort:     getEnvInt("MYSQL_PORT", 3306),
		MySQLUser:     getEnv("MYSQL_USER", "root"),
		MySQLPassword: getEnv("MYSQL_PASSWORD", "123456"),
		MySQLDatabase: getEnv("MYSQL_DATABASE", "noteweb"),

		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 50*1024*1024),

		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
	}
	return cfg
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func (c *Config) DSN() string {
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
