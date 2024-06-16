package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

// handleUnsubscribe handles the /unsubscribe command, allowing the user to remove a subscription.
func (h *BotHandler) handleUnsubscribe(ctx context.Context, chatID int64) error {
	text := "⚠️ An internal error occurred while getting the subscription list. Please try again later."
	subscriptions, err := h.svc.GetAllSubscriptionsByUserID(ctx, chatID)
	if err != nil {
		return h.sendMessage(chatID, text, handleNameListSubscriptions)
	}

	if len(subscriptions) == 0 {
		text = "📋 You have no subscriptions."
		return h.sendMessage(chatID, text, handleNameUnsubscribe)
	}

	var buttons []tgbotapi.InlineKeyboardButton
	for _, sub := range subscriptions {
		buttonText := strings.ReplaceAll(h.buildMessageWithSubscription(sub, false), ", \n", "\n")
		// handleNameUnsubscribe name is used in the callback. TODO: need refactoring
		addPrefixForData := fmt.Sprintf("%s:%s", handleNameUnsubscribe, sub.ID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, addPrefixForData)
		buttons = append(buttons, button)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, button := range buttons {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(button))
	}

	text = "✅ Please choose a subscription to unsubscribe from:"
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err = h.tgBot.SendMessage(msg); err != nil {
		h.l.Error("/unsubscribe: failed to send message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}

// handleUnsubscribeCallback handles the callback query for unsubscribing.
func (h *BotHandler) handleUnsubscribeCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	text := "🟢️ You have been unsubscribed from this subscription.\n\n"
	subscriptionID := callbackQuery.Data
	if err := h.svc.RemoveSubscriptionByID(ctx, subscriptionID); err != nil {
		text = "⚠️ An internal error occurred while unsubscribing. Please try again later."
	}

	if err := h.sendMessage(callbackQuery.Message.Chat.ID, text, handleNameUnsubscribe); err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	// send an updated subscription list.
	return h.handleListSubscriptions(ctx, callbackQuery.Message.Chat.ID)
}
