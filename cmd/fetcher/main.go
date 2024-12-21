package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/gudimz/polovni-auto-alert/internal/app/service/fetcher"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

type Config struct {
	LogLevel string `envconfig:"FETCHER_LOG_LEVEL" default:"info"`
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

	paCliCfg := polovniauto.NewConfig()
	paCli := polovniauto.NewClient(l, paCliCfg)

	svc := fetcher.NewService(l, paCli)

	l.Info("fetcher service started")

	if err := svc.Start(ctx); err != nil {
		l.Error("failed to start fetcher service", logger.ErrAttr(err))
		stop()
	}

	l.Info("fetcher service successfully finished")
}
