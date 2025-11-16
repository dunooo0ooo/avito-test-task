package main

import (
	"context"
	"github.com/dunooo0ooo/avito-test-task/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dunooo0ooo/avito-test-task/internal/app"
	"github.com/dunooo0ooo/avito-test-task/pkg/config"

	userapp "github.com/dunooo0ooo/avito-test-task/internal/user/application"
	userhttp "github.com/dunooo0ooo/avito-test-task/internal/user/delivery/http"
	userpg "github.com/dunooo0ooo/avito-test-task/internal/user/infra/postgres"

	prapp "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/application"
	prhttp "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/delivery/http"
	prpg "github.com/dunooo0ooo/avito-test-task/internal/pullrequest/infra/postgres"

	teamapp "github.com/dunooo0ooo/avito-test-task/internal/team/application"
	teamhttp "github.com/dunooo0ooo/avito-test-task/internal/team/delivery/http"
	teampg "github.com/dunooo0ooo/avito-test-task/internal/team/infra/postgres"

	stats "github.com/dunooo0ooo/avito-test-task/internal/stats/application"
	statshttp "github.com/dunooo0ooo/avito-test-task/internal/stats/delivery/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	log, err := logger.New(cfg.Logger.Level)
	if err != nil {
		log.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer func(log *zap.Logger) {
		_ = log.Sync()
	}(log)

	log.Info("config loaded",
		zap.String("http_addr", cfg.HTTP.Addr),
	)

	poolCfg, err := pgxpool.ParseConfig(cfg.Postgres.DSN())
	if err != nil {
		log.Fatal("cannot parse postgres DSN", zap.Error(err))
	}
	poolCfg.MaxConns = cfg.Postgres.MaxConns
	poolCfg.MinConns = cfg.Postgres.MinConns

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatal("cannot connect to postgres", zap.Error(err))
	}
	defer dbpool.Close()

	log.Info("connected to postgres",
		zap.String("host", cfg.Postgres.Host),
		zap.Int("port", cfg.Postgres.Port),
		zap.String("db", cfg.Postgres.DBName),
	)

	userRepo := userpg.NewUserRepository(dbpool)
	prRepo := prpg.NewPullRequestRepository(dbpool)
	teamRepo := teampg.NewTeamRepository(dbpool)

	userSvc := userapp.NewUserService(userRepo, prRepo, log)
	prSvc := prapp.NewPullRequestService(prRepo, userRepo, log)
	teamSvc := teamapp.NewTeamService(teamRepo, userRepo, log)
	statsSvc := stats.NewStatsService(prRepo, log)

	mux := http.NewServeMux()

	userHandler := userhttp.NewUserHandler(userSvc)
	userHandler.RegisterRoutes(mux)

	prHandler := prhttp.NewPullRequestHandler(prSvc)
	prHandler.RegisterRoutes(mux)

	teamHandler := teamhttp.NewTeamHandler(teamSvc)
	teamHandler.RegisterRoutes(mux)

	statsHandler := statshttp.NewStatsHandler(statsSvc)
	statsHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	application := app.NewApp(srv, log, &cfg)

	if err := application.Start(ctx); err != nil {
		log.Error("application error", zap.Error(err))
	}

	if err := application.Shutdown(); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
}
