package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/telegram"
)

type (
	TgBot interface {
		GetAPI() *tgbotapi.BotAPI
		GetCfg() *telegram.Config
		SendMessage(tgbotapi.Chattable) (tgbotapi.Message, error)
		SetCommands([]tgbotapi.BotCommand) error
	}

	Service interface {
		RemoveAllSubscriptionsByUserID(context.Context, int64 /*userID*/) error
		RemoveSubscriptionByID(context.Context, string /*id*/) error
		CreateSubscription(context.Context, ds.SubscriptionRequest) (ds.SubscriptionResponse, error)
		GetAllSubscriptionsByUserID(context.Context, int64 /*userID*/) ([]ds.SubscriptionResponse, error)
		UpsertUser(context.Context, ds.UserRequest) (ds.UserResponse, error)

		GetCarBrandsList() []string
		GetCarModelsList(string /*brand*/) ([]string, bool)
		GetCarChassisList() map[string]string
		GetRegionsList() map[string]string
	}
)
