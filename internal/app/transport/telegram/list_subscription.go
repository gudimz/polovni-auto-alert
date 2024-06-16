package telegram

import (
	"context"
	"strings"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

// handleListSubscriptions handles the /list_subscriptions command, showing the user their current subscriptions.
func (h *BotHandler) handleListSubscriptions(ctx context.Context, chatID int64) error {
	subscriptions, err := h.svc.GetAllSubscriptionsByUserID(ctx, chatID)
	if err != nil {
		text := "⚠️ An internal error occurred while getting the subscription list. Please try again later."
		return h.sendMessage(chatID, text, handleNameListSubscriptions)
	}

	if len(subscriptions) == 0 {
		text := "📋 You have no subscriptions."
		return h.sendMessage(chatID, text, handleNameListSubscriptions)
	}

	text := h.buildSubscriptionListMessage(subscriptions)

	return h.sendMessage(chatID, text, handleNameListSubscriptions)
}

// buildSubscriptionListMessage builds a message listing all subscriptions.
func (h *BotHandler) buildSubscriptionListMessage(subscriptions []ds.SubscriptionResponse) string {
	var sb strings.Builder

	sb.WriteString("📋 Your current subscriptions:\n")

	for _, sub := range subscriptions {
		sb.WriteString(h.buildMessageWithSubscription(sub, true))
	}

	msg := sb.String()
	msg = strings.ReplaceAll(msg, ", \n", "\n")

	return msg
}
