package main

import (
	"context"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/internal/worker"
	"github.com/ICE-awa/renice-sl/shared/cache"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/database"
	"github.com/ICE-awa/renice-sl/shared/logger"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	defer db.Close()
	slog.Info("PostgreSQL Connected")

	// redis
	rdb, err := cache.NewRedis(ctx, cfg.Redis)
	if err != nil {
		slog.Error("Failed to connect to Redis",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer rdb.Close()
	slog.Info("Redis Connected")

	// nats
	natsClient, err := mq.NewNatsClient(cfg.Nats)
	if err != nil {
		slog.Error("Failed to connect to NATS",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("NATS Connected")

	// jetStream
	err = mq.EnsureStream(natsClient)
	if err != nil {
		slog.Error("Failed to initialize JetStream",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("JetStream Connected")

	// safety browsing
	safeBrowsingClient := util.NewSafeBrowsingClient(cfg.SafeBrowsing.APIKey)

	linkRepo := repository.NewLinkRepository(db)
	dlqRepo := repository.NewDLQRepository(db)

	linkService := service.NewLinkEventService(
		linkRepo,
		dlqRepo,
		rdb,
		safeBrowsingClient,
	)

	linkWorker := worker.NewLinkWorker(linkService, natsClient)
	dlqWorker := worker.NewDLQWorker(natsClient, dlqRepo)

	if err := linkWorker.StartLinkClickWorker(); err != nil {
		slog.Error("Error starting worker",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := linkWorker.StartLinkCheckWorker(); err != nil {
		slog.Error("Error starting worker",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := dlqWorker.StartDLQWorker(); err != nil {
		slog.Error("Error starting DLQ worker",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("All Worker Started")

	r := gin.New()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("worker metrics server listening",
			slog.String("addr", ":8080"),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		slog.Error("Worker metrics server failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)

	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	natsClient.Close()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown worker metrics server",
			slog.String("error", err.Error()),
		)
		_ = srv.Close()
	}

	slog.Info("worker stopped")
}
