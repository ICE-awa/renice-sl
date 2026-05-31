package main

import (
	"context"
	"fmt"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/internal/handler"
	handlerv1 "github.com/ICE-awa/renice-sl/internal/handler/v1"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/internal/router"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/cache"
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
		slog.Error("Error loading config",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Init(cfg.Server.Mode)

	ctx := context.Background()

	// postgres
	db, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		slog.Error("Error initializing postgresql",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// redis
	rdb, err := cache.NewRedis(ctx, cfg.Redis)
	if err != nil {
		slog.Error("Error initializing redis",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer rdb.Close()

	// nats
	natsClient, err := mq.NewNatsClient(cfg.Nats)
	if err != nil {
		slog.Error("Error initializing NATS",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer natsClient.Close()

	// jetstream
	err = mq.EnsureStream(natsClient)
	if err != nil {
		slog.Error("Error initializing JetStream",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 布隆过滤器
	bloom := cache.NewBloomFilter(100_000, 0.01)

	slog.Info("Server Started",
		slog.Int("port", cfg.Server.Port),
		slog.String("mode", cfg.Server.Mode))

	port := fmt.Sprintf(":%d", cfg.Server.Port)

	linkPublisher := event.NewLinkPublisher(natsClient)

	userRepo := repository.NewUserRepository(db)
	linkRepo := repository.NewLinkRepository(db)
	statRepo := repository.NewStatRepository(db)
	dlqRepo := repository.NewDLQRepository(db)

	authSvc := service.NewAuthService(userRepo, rdb, &cfg.Jwt)
	linkSvc := service.NewLinkService(linkRepo, linkPublisher, rdb, &cfg.Link, bloom)
	statSvc := service.NewStatService(statRepo)
	dlqSvc := service.NewDLQService(dlqRepo, natsClient)

	// 布隆过滤器初始化
	err = linkSvc.InitBloomFilter()
	if err != nil {
		slog.Error("Error initializing Bloom Filter",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	h := &handler.Handlers{
		HealthH: handler.NewHealthHandler(db, rdb, natsClient),
		AuthHV1: handlerv1.NewAuthHandler(authSvc, &cfg.Jwt),
		LinkHV1: handlerv1.NewLinkHandler(linkSvc),
		StatHV1: handlerv1.NewStatHandler(statSvc),
		DLQHV1:  handlerv1.NewDLQHandler(dlqSvc),
	}

	r := router.Setup(h, &cfg.Jwt, rdb, userRepo)

	err = r.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	})
	if err != nil {
		panic(err)
	}

	if err := r.Run(port); err != nil {
		slog.Error("Error occurred while server starting",
			slog.Int("port", cfg.Server.Port),
			slog.String("mode", cfg.Server.Mode),
			slog.String("error", err.Error()))
	}

}

//func startPGPoolStatsLogger(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
//	ticker := time.NewTicker(interval)
//
//	go func() {
//		defer ticker.Stop()
//
//		for {
//			select {
//			case <-ctx.Done():
//				return
//
//			case <-ticker.C:
//				s := pool.Stat()
//
//				log.Printf(
//					"pgpool max=%d total=%d acquired=%d idle=%d constructing=%d acquire_count=%d empty_acquire=%d acquire_duration=%s canceled_acquire=%d",
//					s.MaxConns(),
//					s.TotalConns(),
//					s.AcquiredConns(),
//					s.IdleConns(),
//					s.ConstructingConns(),
//					s.AcquireCount(),
//					s.EmptyAcquireCount(),
//					s.AcquireDuration(),
//					s.CanceledAcquireCount(),
//				)
//			}
//		}
//	}()
//}
