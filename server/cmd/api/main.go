package main

import (
	"context"
	"fmt"
	"github.com/ICE-awa/renice-sl/internal/handler"
	"github.com/ICE-awa/renice-sl/internal/router"
	"github.com/ICE-awa/renice-sl/shared/cache"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/database"
	"github.com/ICE-awa/renice-sl/shared/logger"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"log/slog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Error loading config",
			slog.String("error", err.Error()))
		return
	}

	logger.Init(cfg.Server.Mode)

	ctx := context.Background()

	db, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		slog.Error("Error initializing postgresql",
			slog.String("error", err.Error()))
	}
	defer db.Close()

	rdb, err := cache.NewRedis(ctx, cfg.Redis)
	if err != nil {
		slog.Error("Error initializing redis",
			slog.String("error", err.Error()))
	}
	defer rdb.Close()

	natsClient, err := mq.NewNatsClient(cfg.Nats)
	if err != nil {
		slog.Error("Error initializing NATS",
			slog.String("error", err.Error()))
	}
	defer natsClient.Close()

	slog.Info("Server Started",
		slog.Int("port", cfg.Server.Port),
		slog.String("mode", cfg.Server.Mode))

	port := fmt.Sprintf(":%d", cfg.Server.Port)

	h := &handler.Handlers{
		HealthH: handler.NewHealthHandler(db, rdb, natsClient),
	}

	r := router.Setup(h)

	if err := r.Run(port); err != nil {
		slog.Error("Error occurred while server starting",
			slog.Int("port", cfg.Server.Port),
			slog.String("mode", cfg.Server.Mode),
			slog.String("error", err.Error()))
	}

}
