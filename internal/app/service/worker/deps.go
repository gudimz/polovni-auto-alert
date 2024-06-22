package worker

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/telegram"
)

//go:generate mockgen -source=deps.go -destination=deps_mock.go -package=worker
type (
	Repository interface {
		UpsertListing(ctx context.Context, listing ds.UpsertListingRequest) error
		GetListingsByIsNeedSend(ctx context.Context, isNeedSend bool) ([]ds.ListingResponse, error)
		CreateNotification(ctx context.Context, notification ds.CreateNotificationRequest) (ds.NotificationResponse, error)
		GetSubscriptionByID(ctx context.Context, id string) (ds.SubscriptionResponse, error)
		GetSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error)
		DeleteListingsBySubscriptionIDs(ctx context.Context, ids []string) error
		DeleteSubscriptionsByUserID(ctx context.Context, userID int64) error
		DeleteUserByID(ctx context.Context, id int64) error
	}
	TgBot interface {
		GetAPI() *tgbotapi.BotAPI
		GetCfg() *telegram.Config
		SendMessage(c tgbotapi.Chattable) (tgbotapi.Message, error)
		SetCommands(commands []tgbotapi.BotCommand) error
	}
)
