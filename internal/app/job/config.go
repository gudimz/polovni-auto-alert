package job

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Interval time.Duration `envconfig:"PA_JOB_INTERVAL" default:"24h"`
}

func NewConfig() *Config {
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	return cfg
}
