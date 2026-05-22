package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"my-blog/backend/internal/config"
	"my-blog/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	migrationsDir := filepath.Join("migrations")
	if err := db.RunMigrations(cfg.DatabaseURL, migrationsDir); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations completed")
}
