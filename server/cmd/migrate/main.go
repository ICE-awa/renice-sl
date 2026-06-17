package main

import (
	"errors"
	migrations "github.com/ICE-awa/renice-sl/migrations"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"log/slog"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Init(cfg.Server.Mode)

	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		slog.Error("failed to create migration source",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		sourceDriver,
		cfg.Database.DSN(),
	)
	if err != nil {
		slog.Error("failed to create migrator",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("database migration no change")
			return
		}

		slog.Error("failed to run migration",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
}
