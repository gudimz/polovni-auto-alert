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
		SendMessage(c tgbotapi.Chattable) (tgbotapi.Message, error)
		SetCommands(commands []tgbotapi.BotCommand) error
	}

	Service interface {
		RemoveAllSubscriptionsByUserID(ctx context.Context, userID int64) error
		RemoveSubscriptionByID(ctx context.Context, id string) error
		CreateSubscription(ctx context.Context, subscription ds.SubscriptionRequest) (ds.SubscriptionResponse, error)
		GetAllSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error)
		UpsertUser(ctx context.Context, user ds.UserRequest) (ds.UserResponse, error)

		GetCarsList() map[string][]string
		GetChassisList() map[string]string
		GetRegionsList() map[string]string
	}
)
