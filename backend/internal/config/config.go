package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppName            string
	HTTPPort           string
	DatabaseURL        string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	CORSAllowedOrigins []string
	StorageDriver      string
	StorageLocalDir    string
	StoragePublicURL   string
	UploadMaxBytes     int64
	WorkerPollInterval time.Duration
	WorkerBatchSize    int
	WorkerMaxAttempts  int
	WorkerRetryDelay   time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppName:            getEnv("APP_NAME", "umkm-commerce-os"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:          strings.TrimSpace(os.Getenv("JWT_SECRET")),
		AccessTokenTTL:     time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL:    time.Duration(getEnvInt("REFRESH_TOKEN_TTL_DAYS", 30)) * 24 * time.Hour,
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		StorageDriver:      getEnv("STORAGE_DRIVER", "local"),
		StorageLocalDir:    getEnv("STORAGE_LOCAL_DIR", "uploads"),
		StoragePublicURL:   strings.TrimRight(getEnv("STORAGE_PUBLIC_URL", "http://localhost:8080/uploads"), "/"),
		UploadMaxBytes:     int64(getEnvInt("UPLOAD_MAX_BYTES", 5*1024*1024)),
		WorkerPollInterval: time.Duration(getEnvInt("OUTBOX_POLL_INTERVAL_SECONDS", 5)) * time.Second,
		WorkerBatchSize:    getEnvInt("OUTBOX_BATCH_SIZE", 10),
		WorkerMaxAttempts:  getEnvInt("OUTBOX_MAX_ATTEMPTS", 5),
		WorkerRetryDelay:   time.Duration(getEnvInt("OUTBOX_RETRY_DELAY_SECONDS", 30)) * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}

	port, err := strconv.Atoi(c.HTTPPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("HTTP_PORT must be a valid TCP port")
	}
	if c.AccessTokenTTL <= 0 {
		return errors.New("ACCESS_TOKEN_TTL_MINUTES must be greater than zero")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("REFRESH_TOKEN_TTL_DAYS must be greater than zero")
	}
	if c.StorageDriver != "local" {
		return errors.New("STORAGE_DRIVER must be local for the current backend foundation")
	}
	if strings.TrimSpace(c.StorageLocalDir) == "" {
		return errors.New("STORAGE_LOCAL_DIR is required")
	}
	if strings.TrimSpace(c.StoragePublicURL) == "" {
		return errors.New("STORAGE_PUBLIC_URL is required")
	}
	if c.UploadMaxBytes <= 0 {
		return errors.New("UPLOAD_MAX_BYTES must be greater than zero")
	}
	if c.WorkerPollInterval <= 0 {
		return errors.New("OUTBOX_POLL_INTERVAL_SECONDS must be greater than zero")
	}
	if c.WorkerBatchSize <= 0 {
		return errors.New("OUTBOX_BATCH_SIZE must be greater than zero")
	}
	if c.WorkerMaxAttempts <= 0 {
		return errors.New("OUTBOX_MAX_ATTEMPTS must be greater than zero")
	}
	if c.WorkerRetryDelay <= 0 {
		return errors.New("OUTBOX_RETRY_DELAY_SECONDS must be greater than zero")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
