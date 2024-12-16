package polovniauto

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PageLimit     int           `envconfig:"PAGE_LIMIT" default:"9999"`
	ChromeWSURL   string        `envconfig:"CHROME_WS_URL" default:"ws://chrome:3000"`
	ChromeTimeout time.Duration `envconfig:"CHROME_TIMEOUT" default:"10m"`
}

func NewConfig() *Config {
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	return cfg
}
