package job

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type Job struct {
	l        *logger.Logger
	cfg      *Config
	notifier Notifier
	scrapper Scrapper
}

const (
	NotifierCarsListJobName = "update_notifier_cars_list_job"
	NotifierRegionsJobName  = "update_notifier_regions_job"
	NotifierChassisJobName  = "update_notifier_chassis_job"
	ScrapperCarsListJobName = "update_scrapper_cars_list_job"
)

func New(l *logger.Logger, cfg *Config, notifier Notifier, scrapper Scrapper) *Job {
	return &Job{
		l:        l,
		cfg:      cfg,
		notifier: notifier,
		scrapper: scrapper,
	}
}

// Execute runs the jobs in the background.
func (j *Job) Execute(ctx context.Context, jobs map[string]func(context.Context) error) error {
	defer j.recoverPanic()

	done := make(chan struct{})
	defer close(done)

	for name, job := range jobs {
		go func(name string, job func(context.Context) error) {
			ticker := time.NewTicker(j.cfg.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					j.l.Info("Job stopped", logger.StringAttr("job_name", name))
					return
				case <-ticker.C:
					j.l.Info("Running job", logger.StringAttr("job_name", name))

					if err := job(ctx); err != nil {
						j.l.Error("Job error", logger.ErrAttr(err), logger.StringAttr("job_name", name))
					} else {
						j.l.Info("Job completed successfully", logger.StringAttr("job_name", name))
					}
				}
			}
		}(name, job)
	}

	// wait for the context to be done
	<-ctx.Done()
	j.l.Info("All jobs stopped")

	return nil
}

// recoverPanic recovers from a panic and logs the error.
func (j *Job) recoverPanic() {
	if r := recover(); r != nil {
		j.l.Error(
			"Recovered from panic",
			logger.AnyAttr("error", r),
			logger.StringAttr("stacktrace", string(debug.Stack())),
		)
	}
}
