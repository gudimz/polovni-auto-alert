package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db"
	"github.com/gudimz/polovni-auto-alert/internal/app/service/notifier"
	"github.com/gudimz/polovni-auto-alert/internal/app/transport/telegram"
	"github.com/gudimz/polovni-auto-alert/internal/pkg/data"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	tgCli "github.com/gudimz/polovni-auto-alert/pkg/telegram"
)

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
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

	dbCfg := db.NewConfig()

	repo, err := db.NewRepo(ctx, l, dbCfg)
	if err != nil {
		l.Error("failed to initialize repository", logger.ErrAttr(err))
		return
	}
	defer repo.Close()

	if err = repo.Migrate(); err != nil {
		l.Error("failed to migrate database", logger.ErrAttr(err))
		return
	}

	tgCfg := tgCli.NewConfig()

	bot, err := tgCli.NewBot(l, tgCfg)
	if err != nil {
		l.Error("failed to create bot", logger.ErrAttr(err))
		return
	}

	dataLoader, err := data.NewLoader()
	if err != nil {
		l.Error("failed to create data loader", logger.ErrAttr(err))
	}

	svc := notifier.NewService(
		l,
		repo,
		dataLoader.GetCarsList(),
		dataLoader.GetChassisList(),
		dataLoader.GetRegionsList(),
	)

	tgHandler := telegram.NewBotHandler(l, bot, svc)

	go func() {
		if err = tgHandler.Start(ctx); err != nil {
			l.Error("failed to start notifier service", logger.ErrAttr(err))
			stop()
		}
	}()

	l.Info("notifier service started")

	<-ctx.Done()

	l.Info("notifier service shutting down")

	stop()

	l.Info("service stopped gracefully")
}
