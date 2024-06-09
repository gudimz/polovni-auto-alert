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

func run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	l := logger.NewLogger(
		logger.WithLevel(cfg.LogLevel),
		logger.WithAddSource(true),
		logger.WithIsJSON(true))
	logger.SetDefault(l)

	dbCfg := db.NewConfig()
	repo, err := db.NewRepo(ctx, l, dbCfg)
	if err != nil {
		l.Error("failed to initialize repository", logger.ErrAttr(err))
		return
	}
	defer repo.Close()

	tgCfg := tgCli.NewConfig()
	bot, err := tgCli.NewBot(l, tgCfg)
	if err != nil {
		l.Error("failed to create bot", logger.ErrAttr(err))
		return
	}

	svc := worker.NewService(l, repo, bot, cfg.WorkerNotificationInterval)

	go func() {
		if err = svc.Start(ctx); err != nil {
			l.Error("failed to start worker service", logger.ErrAttr(err))
			stop()
		}
	}()

	l.Info("worker service started")

	<-ctx.Done()

	stop()

	l.Info("service stopped gracefully")
}
