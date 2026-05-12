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
)

func main() {
	// config
	cfg, _ := config.Load()

	// slog
	logger.Init(cfg.Server.Mode)

	ctx := context.Background()

	// database
	db, _ := database.NewPool(ctx, cfg.Database)

	// nats
	natsClient, _ := mq.NewNatsClient(cfg.Nats)
	defer natsClient.Close()

	// jetStream
	_ = mq.EnsureStream(natsClient)

	linkRepo := repository.NewLinkRepository(db)

	linkService := service.NewLinkEventService(linkRepo)

	linkWorker := worker.NewLinkClickWorker(linkService)

	if err := linkWorker.StartLinkClickWorker(natsClient); err != nil {
		slog.Error("Error starting worker",
			slog.String("error", err.Error()))
	}

	select {}
}
