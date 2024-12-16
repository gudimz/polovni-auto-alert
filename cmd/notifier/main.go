package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"github.com/gudimz/polovni-auto-alert/internal/app/job"
	"github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db"
	"github.com/gudimz/polovni-auto-alert/internal/app/service/notifier"
	"github.com/gudimz/polovni-auto-alert/internal/app/transport/telegram"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
	tgCli "github.com/gudimz/polovni-auto-alert/pkg/telegram"
)

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

func main() {
	run()
}

// run starts the notifier service
//
//nolint:funlen
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

	bot, err := tgCli.NewBot(l, tgCli.NewConfig())
	if err != nil {
		l.Error("failed to create bot", logger.ErrAttr(err))
		return
	}

	paCli := polovniauto.NewClient(l, polovniauto.NewConfig())

	svc := notifier.NewService(
		l,
		repo,
		paCli,
	)

	// run critical jobs before starting the handler
	jobsList := map[string]func(context.Context) error{
		job.NotifierCarsListJobName: svc.UpdateCarList,
		job.NotifierRegionsJobName:  svc.UpdateCarRegionsList,
		job.NotifierChassisJobName:  svc.UpdateCarChassisList,
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

	l.Info("critical jobs  for notifier completed successfully")

	jobs := job.New(l, job.NewConfig(), svc, nil)

	go func() {
		if err = jobs.Execute(ctx, jobsList); err != nil {
			l.Error("failed to start jobs for notifier", logger.ErrAttr(err))
			stop()
		}
	}()

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
