package main

import (
	"context"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/internal/worker"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/database"
	"github.com/ICE-awa/renice-sl/shared/logger"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"log/slog"
	"os"
)

func main() {
	// config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// slog
	logger.Init(cfg.Server.Mode)

	ctx := context.Background()

	// database
	db, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// nats
	natsClient, err := mq.NewNatsClient(cfg.Nats)
	defer natsClient.Close()
	if err != nil {
		slog.Error("Failed to connect to NATS",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// jetStream
	err = mq.EnsureStream(natsClient)
	if err != nil {
		slog.Error("Failed to initialize JetStream",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	linkRepo := repository.NewLinkRepository(db)

	linkService := service.NewLinkEventService(linkRepo)

	linkWorker := worker.NewLinkClickWorker(linkService)

	if err := linkWorker.StartLinkClickWorker(natsClient); err != nil {
		slog.Error("Error starting worker",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	select {}
}
