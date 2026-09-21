package main

import (
	"astrohop/internal/account"
	"astrohop/internal/infrastructure/mem"
	"astrohop/internal/infrastructure/postgres"
	"astrohop/internal/mission"
	"astrohop/internal/pairing"
	"astrohop/internal/search"
	"astrohop/internal/web"
	"astrohop/internal/web/middleware"
	"astrohop/pkg/config"
	"astrohop/pkg/db"
	"astrohop/pkg/logger"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := logger.GetLogger(os.Stdout)
	cfg := config.GetConfig()

	infra, err := initInfra(ctx, cfg)
	if err != nil {
		log.Error("failed to initialize infrastructure", "cause", err.Error())
		os.Exit(1)
	}

	searchRepo := postgres.NewSearchRepo(infra.manager)
	searchSvc := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchSvc)
	missionRepo := postgres.NewMissionRepo(infra.manager)
	taskQueue := mem.NewTaskQueue()
	progressHub := mem.NewProgressHub()
	missionSvc := mission.NewService(missionRepo, taskQueue, progressHub, searchSvc, log)
	missionSvc.Start(ctx)
	missionHandler := mission.NewHandler(missionSvc)
	accountRepo := postgres.NewAccountRepo(infra.manager)
	accountSvc := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountSvc)
	sessionRegistry := mem.NewSessionRegistry(log)
	pairingSvc := pairing.NewService(sessionRegistry)
	pairingHandler := pairing.NewHandler(pairingSvc)
	authMware := middleware.NewAuth(accountRepo)
	errorMware := middleware.NewError(log)

	server := web.NewServer(
		accountHandler,
		searchHandler,
		missionHandler,
		pairingHandler,
		authMware,
		errorMware,
	)

	srv := server.Server("0.0.0.0:8080")
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen failed", "cause", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Error("server shutdown failed", "cause", err.Error())
		os.Exit(1)
	}

	infra.manager.Pool.Close()
	log.Info("server exiting")
}

type infrastructure struct {
	manager *db.Manager
}

func initInfra(ctx context.Context, cfg *config.Config) (*infrastructure, error) {
	var infra infrastructure

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
