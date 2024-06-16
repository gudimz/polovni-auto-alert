package telegram

import "context"

// handleStop handles the /stop command, removing the user's subscriptions.
func (h *BotHandler) handleStop(ctx context.Context, chatID int64) error {
	text := "🟢️ You have been unsubscribed from all notifications."
	if err := h.svc.RemoveAllSubscriptionsByUserID(ctx, chatID); err != nil {
		text = "⚠️ An internal error occurred while unsubscribing. Please try again later."
	}

	return h.sendMessage(chatID, text, handleNameStop)
}
