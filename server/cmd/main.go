package main

import (
	"astrohop/internal/account"
	"astrohop/internal/infrastructure/postgres"
	"astrohop/internal/mission"
	"astrohop/internal/search"
	"astrohop/internal/web"
	"astrohop/pkg/config"
	"astrohop/pkg/db"
	"astrohop/pkg/logger"
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.L().Infow("Starting app")
	infra, err := initInfra(ctx)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initialize infrastructure: %w", err))
	}

	searchRepo := postgres.NewSearchRepo(infra.manager)
	searchSvc := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchSvc)
	missionRepo := postgres.NewMissionRepo(infra.manager)
	missionSvc := mission.NewService(missionRepo, searchSvc)
	missionSvc.Start(ctx)
	missionHandler := mission.NewHandler(missionSvc)
	accountRepo := postgres.NewAccountRepo(infra.manager)
	accountSvc := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountSvc)

	server := web.NewServer(
		missionRepo,
		accountRepo,
		accountHandler,
		searchHandler,
		missionHandler,
	)

	select {
	case <-ctx.Done():
		logger.L().Infow("Shutting down server")
		return
	default:
		logger.L().Infow("Starting HTTP server")
		logger.L().Fatal(server.Start("0.0.0.0:8080"))
	}
}

type infrastructure struct {
	manager *db.Manager
}

func initInfra(ctx context.Context) (*infrastructure, error) {
	var infra infrastructure
	cfg := config.C()
	pool, err := pgxpool.New(ctx, cfg.Infra.DBConnectionString)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	infra.manager = db.NewManager(pool)
	return &infra, nil
}
