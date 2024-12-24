package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db"
	"github.com/gudimz/polovni-auto-alert/internal/app/service/fetcher"
	"github.com/gudimz/polovni-auto-alert/internal/app/service/scraper"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

type Config struct {
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	ScraperInterval time.Duration `envconfig:"SCRAPER_INTERVAL" default:"10m"`
	Workers         int           `envconfig:"SCRAPER_WORKERS_COUNT" default:"5"`
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

	paCliCfg := polovniauto.NewConfig()
	paCli := polovniauto.NewClient(l, paCliCfg)

	fetch := fetcher.NewService(l, paCli)

	svc := scraper.NewService(
		l,
		repo,
		paCli,
		fetch,
		cfg.ScraperInterval,
		cfg.Workers,
	)

	go func() {
		if err = svc.Start(ctx); err != nil {
			l.Error("failed to start scraper service", logger.ErrAttr(err))
			stop()
		}
	}()

	l.Info("scraper service started")

	<-ctx.Done()

	l.Info("scraper service shutting down")

	stop()

	l.Info("scraper service stopped gracefully")
}
