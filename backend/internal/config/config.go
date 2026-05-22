package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	AppPort             string
	DatabaseURL         string
	JWTAccessSecret     string
	JWTRefreshSecret    string
	JWTAccessExpires    time.Duration
	JWTRefreshExpires   time.Duration
	UploadDir           string
	PublicUploadBaseURL string
	MaxImageSizeMB      int64
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		AppPort:             getEnv("APP_PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://myblog:myblog@localhost:5432/myblog?sslmode=disable"),
		JWTAccessSecret:     getEnv("JWT_ACCESS_SECRET", "change-me-access-secret-min-32-chars"),
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", "change-me-refresh-secret-min-32-chars"),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		PublicUploadBaseURL: getEnv("PUBLIC_UPLOAD_BASE_URL", "/uploads"),
		MaxImageSizeMB:      getEnvInt64("MAX_IMAGE_SIZE_MB", 5),
	}

	var err error
	cfg.JWTAccessExpires, err = time.ParseDuration(getEnv("JWT_ACCESS_EXPIRES", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRES: %w", err)
	}
	cfg.JWTRefreshExpires, err = time.ParseDuration(getEnv("JWT_REFRESH_EXPIRES", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRES: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}
