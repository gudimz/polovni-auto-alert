package telegram

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type Bot struct {
	l   *logger.Logger
	cfg *Config
	API *tgbotapi.BotAPI
}

// NewBot creates a new Bot instance with the configuration.
func NewBot(l *logger.Logger, cfg *Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, errors.Wrap(err, "new bot failed")
	}

	bot.Debug = cfg.IsDebug

	l.Info("Authorized on account", logger.StringAttr("username", bot.Self.UserName))

	return &Bot{
		l:   l,
		cfg: cfg,
		API: bot,
	}, nil
}

// GetAPI returns the bot's API instance.
func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.API
}

// GetCfg returns the bot's configuration.
func (b *Bot) GetCfg() *Config {
	return b.cfg
}

// SendMessage sends a message using the bot's API.
func (b *Bot) SendMessage(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.API.Send(c) //nolint:wrapcheck,nolintlint
}

// SetCommands sets the bot commands that will be shown in the UI.
func (b *Bot) SetCommands(commands []tgbotapi.BotCommand) error {
	cfg := tgbotapi.NewSetMyCommands(commands...)

	_, err := b.API.Request(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to set commands")
	}

	return nil
}
