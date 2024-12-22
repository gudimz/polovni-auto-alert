package telegram

import (
	"context"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

const startButtonsPerRow = 2

// handleStart handles the /start command.
func (h *BotHandler) handleStart(ctx context.Context, chat tgbotapi.Chat) error {
	text := `
👋 Welcome to Polovni Automobili Alert Bot!

This bot helps you stay updated with the latest car listings that match your preferences.

Here’s what you can do:

📬 subscribe - Subscribe to new car listings alerts
❌ unsubscribe - Unsubscribe from a car listings alert
📋 list_subscriptions - List all your current subscriptions
🚫 stop - Stop receiving notifications

Just select the desired command or type it in the chat to get started.
`
	_, err := h.svc.UpsertUser(ctx, ds.UserRequest{
		ID:        chat.ID,
		Username:  chat.UserName,
		FirstName: chat.FirstName,
		LastName:  chat.LastName,
	})

	if err != nil {
		text = "⚠️ An internal error occurred while starting. Please try again later."
	}

	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📬 Subscribe", handleNameSubscribe),
		tgbotapi.NewInlineKeyboardButtonData("❌ Unsubscribe", handleNameUnsubscribe),
		tgbotapi.NewInlineKeyboardButtonData("📋 List Subscriptions", handleNameListSubscriptions),
		tgbotapi.NewInlineKeyboardButtonData("🚫 Stop", handleNameStop),
	}

	keyboard := createKeyboard(ctx, startButtonsPerRow, nil, buttons)

	msg := tgbotapi.NewMessage(chat.ID, text)
	msg.ReplyMarkup = keyboard

	if _, err = h.tgBot.SendMessage(msg); err != nil {
		h.l.Error("/start: failed to send message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}
