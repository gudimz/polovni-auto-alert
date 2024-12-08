package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/gudimz/polovni-auto-alert/internal/app/service/worker"
	tgCli "github.com/gudimz/polovni-auto-alert/pkg/telegram"

	"github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type Config struct {
	LogLevel                   string        `envconfig:"LOG_LEVEL" default:"info"`
	WorkerNotificationInterval time.Duration `envconfig:"WORKER_NOTIFICATION_INTERVAL" default:"20m"`
}

func main() {
	run()
}

// run starts the worker service.
func run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	lg := logger.NewLogger(
		logger.WithLevel(cfg.LogLevel),
		logger.WithAddSource(true),
		logger.WithIsJSON(true))
	logger.SetDefault(lg)

	dbCfg := db.NewConfig()

	repo, err := db.NewRepo(ctx, lg, dbCfg)
	if err != nil {
		lg.Error("failed to initialize repository", logger.ErrAttr(err))
		return
	}
	defer repo.Close()

	tgCfg := tgCli.NewConfig()

	bot, err := tgCli.NewBot(lg, tgCfg)
	if err != nil {
		lg.Error("failed to create bot", logger.ErrAttr(err))
		return
	}

	svc := worker.NewService(lg, repo, bot, cfg.WorkerNotificationInterval)

	go func() {
		if err = svc.Start(ctx); err != nil {
			lg.Error("failed to start worker service", logger.ErrAttr(err))
			stop()
		}
	}()

	lg.Info("worker service started")

	<-ctx.Done()

	stop()

	lg.Info("service stopped gracefully")
}
