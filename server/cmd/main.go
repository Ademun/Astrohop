package main

import (
	"astrohop/internal/content"
	"astrohop/internal/infrastructure/postgres"
	"astrohop/internal/planner"
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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.L().Infow("Starting app")
	infra, err := initInfra(ctx)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to initialize infrastructure: %w", err))
	}

	h := initHandlers(ctx, infra)
	server := web.NewServer(
		h.searchHandler,
		h.plannerHandler,
		h.contentHandler,
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
	manager     *db.Manager
	minioClient *minio.Client
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
	minioClient, err := minio.New(config.C().Infra.MinIOEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.C().Infra.MinIORootUser,
			config.C().Infra.MinIORootPassword,
			"",
		),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	infra.minioClient = minioClient
	return &infra, nil
}

type handlers struct {
	searchHandler  *search.Handler
	plannerHandler *planner.Handler
	contentHandler *content.Handler
}

func initHandlers(ctx context.Context, infra *infrastructure) *handlers {
	var h handlers
	searchRepo := postgres.NewSearchRepo(infra.manager)
	searchSvc := search.NewService(searchRepo)
	h.searchHandler = search.NewHandler(searchSvc)
	missionRepo := postgres.NewMissionRepo(infra.manager)
	contentSvc := content.NewService(infra.minioClient)
	h.contentHandler = content.NewHandler(contentSvc)
	plannerSvc := planner.NewService(missionRepo, searchSvc, contentSvc, infra.minioClient)
	plannerSvc.Start(ctx)
	h.plannerHandler = planner.NewHandler(plannerSvc)
	return &h
}
