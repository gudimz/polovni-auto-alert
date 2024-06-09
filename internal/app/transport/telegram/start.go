package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

// handleStart handles the /start command.
func (h *BotHandler) handleStart(ctx context.Context, chat *tgbotapi.Chat) error {
	text := `
👋 Welcome to Polovni Automobili Alert Bot!

This bot helps you stay updated with the latest car listings that match your preferences.

Here’s what you can do:

/subscribe - 📬 Subscribe to new car listings alerts
/unsubscribe - ❌ Unsubscribe from a car listings alert
/list_subscriptions - 📋 List all your current subscriptions
/stop - 🚫 Stop receiving notifications

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

	return h.sendMessage(chat.ID, text, handleNameStart)
}
