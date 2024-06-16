package telegram

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds the configuration for the Telegram bot.
type Config struct {
	BotToken         string `envconfig:"TELEGRAM_API_TOKEN" required:"true"`
	UpdateCfgTimeout int    `envconfig:"TELEGRAM_UPDATE_CONFIG_TIMEOUT" default:"60"`
	IsDebug          bool   `envconfig:"TELEGRAM_DEBUG" default:"false"`
}

func NewConfig() *Config {
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	return cfg
}
