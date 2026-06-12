package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPass            string
	SMTPFrom            string
	NtfyBaseURL         string
	NtfyInternalURL     string
	NtfyUsername        string
	NtfyPassword        string
	NtfyAccessToken     string
	NtfyCLIPath         string
	NtfyConfigFile      string
	NtfyDockerContainer string
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
		SMTPHost:            getEnv("SMTP_HOST", ""),
		SMTPPort:            getEnv("SMTP_PORT", "587"),
		SMTPUser:            getEnv("SMTP_USER", ""),
		SMTPPass:            getEnv("SMTP_PASS", ""),
		SMTPFrom:            getEnv("SMTP_FROM", ""),
		NtfyBaseURL:         getEnv("NTFY_BASE_URL", ""),
		NtfyInternalURL:     getEnv("NTFY_INTERNAL_URL", ""),
		NtfyUsername:        getEnv("NTFY_USERNAME", "blog"),
		NtfyPassword:        getEnv("NTFY_PASSWORD", ""),
		NtfyAccessToken:     getEnv("NTFY_ACCESS_TOKEN", ""),
		NtfyCLIPath:         getEnv("NTFY_CLI_PATH", "ntfy"),
		NtfyConfigFile:      getEnv("NTFY_CONFIG_FILE", ""),
		NtfyDockerContainer: getEnv("NTFY_DOCKER_CONTAINER", ""),
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
	if strings.EqualFold(cfg.AppEnv, "production") {
		if err := validateProductionSecret("JWT_ACCESS_SECRET", cfg.JWTAccessSecret); err != nil {
			return nil, err
		}
		if err := validateProductionSecret("JWT_REFRESH_SECRET", cfg.JWTRefreshSecret); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func validateProductionSecret(name, value string) error {
	if len(value) < 32 {
		return fmt.Errorf("%s must be at least 32 characters in production", name)
	}
	defaults := map[string]struct{}{
		"change-me-access-secret-min-32-chars":  {},
		"change-me-refresh-secret-min-32-chars": {},
	}
	if _, ok := defaults[value]; ok {
		return fmt.Errorf("%s must not use the default value in production", name)
	}
	if strings.Contains(strings.ToLower(value), "change-me") {
		return fmt.Errorf("%s must not use a placeholder value in production", name)
	}
	return nil
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
