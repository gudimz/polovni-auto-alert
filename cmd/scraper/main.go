package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/gudimz/polovni-auto-alert/internal/app/job"
	"github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db"
	"github.com/gudimz/polovni-auto-alert/internal/app/service/scraper"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

type Config struct {
	LogLevel           string        `envconfig:"LOG_LEVEL" default:"info"`
	ScraperInterval    time.Duration `envconfig:"SCRAPER_INTERVAL" default:"10m"`
	ScraperStartOffset time.Duration `envconfig:"SCRAPER_START_OFFSET" default:"0m"`
	Workers            int           `envconfig:"SCRAPER_WORKERS_COUNT" default:"5"`
}

func main() {
	run()
}

// run starts the scraper service.
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

	repo, err := db.NewRepo(ctx, l, db.NewConfig())
	if err != nil {
		l.Error("failed to initialize repository", logger.ErrAttr(err))
		return
	}
	defer repo.Close()

	if err = repo.Migrate(); err != nil {
		l.Error("failed to migrate database", logger.ErrAttr(err))
		return
	}

	paCli := polovniauto.NewClient(l, polovniauto.NewConfig())

	svc := scraper.NewService(
		l,
		repo,
		paCli,
		cfg.ScraperInterval,
		cfg.ScraperStartOffset,
		cfg.Workers,
	)

	// run critical jobs before starting the service
	jobsList := map[string]func(context.Context) error{
		job.ScrapperCarsListJobName: svc.UpdateCarChassisList,
	}

	var wg sync.WaitGroup

	errCh := make(chan error, len(jobsList))

	for name, f := range jobsList {
		wg.Add(1)

		go func(name string, f func(context.Context) error) {
			defer wg.Done()

			l.Info("running critical job before start", logger.StringAttr("job_name", name))

			if errF := f(ctx); errF != nil {
				l.Error("failed to run critical job before start", logger.ErrAttr(errF), logger.StringAttr("job_name", name))
				errCh <- errF
			}
		}(name, f)
	}

	wg.Wait()
	close(errCh)

	for err = range errCh {
		if err != nil {
			stop()
			return
		}
	}

	l.Info("critical jobs for scrapper completed successfully")

	go func() {
		if err = svc.Start(ctx); err != nil {
			l.Error("failed to start scraper service", logger.ErrAttr(err))
			stop()
		}
	}()

	jobCfg := job.NewConfig()
	jobs := job.New(l, jobCfg, nil, svc)

	go func() {
		if err = jobs.Execute(ctx, jobsList); err != nil {
			l.Error("failed to start jobs for scrapper", logger.ErrAttr(err))
			stop()
		}
	}()

	l.Info("scraper service started")

	<-ctx.Done()

	l.Info("scraper service shutting down")

	stop()

	l.Info("service stopped gracefully")
}
