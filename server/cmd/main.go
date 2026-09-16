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
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := logger.GetLogger()
	cfg := config.GetConfig()

	infra, err := initInfra(ctx, cfg)
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
	sessionRegistry := mem.NewSessionRegistry(log)
	pairingSvc := pairing.NewService(sessionRegistry)
	pairingHandler := pairing.NewHandler(pairingSvc)
	authMware := middleware.NewAuth(accountRepo)
	permViewMware := middleware.NewPermissions(accountRepo, middleware.ActionView)
	permEditMware := middleware.NewPermissions(accountRepo, middleware.ActionEdit)
	permDeleteMware := middleware.NewPermissions(accountRepo, middleware.ActionDelete)
	errorMware := middleware.NewError(log)

	server := web.NewServer(
		accountHandler,
		searchHandler,
		missionHandler,
		pairingHandler,
		authMware,
		permViewMware,
		permEditMware,
		permDeleteMware,
		errorMware,
	)

	srv := server.Server("0.0.0.0:8080")
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down...")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
	infra.manager.Pool.Close()
	log.Info("Server exiting")
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
