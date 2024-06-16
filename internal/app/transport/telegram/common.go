package telegram

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

// sendMessage sends a text message to the user.
func (h *BotHandler) sendMessage(chatID int64, text, handlerName string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error(handlerName+": failed to send message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}

// sendUnknownCommandMessage sends a message indicating an unknown command.
func (h *BotHandler) sendUnknownCommandMessage(_ context.Context, chatID int64) error {
	text := "🤔 I'm not sure what you mean. Please use one of the available commands."
	return h.sendMessage(chatID, text, handleNameUnknown)
}

// generateButtons generates inline keyboard buttons from a map of items.
func generateButtons(_ context.Context, items map[string]string) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton
	for item := range items {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(item, item))
	}

	sort.Slice(buttons, func(i, j int) bool {
		return buttons[i].Text < buttons[j].Text
	})

	return buttons
}

// generateButtonsFromSlice generates inline keyboard buttons from a slice of items.
func generateButtonsFromSlice(_ context.Context, items []string) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton
	for _, item := range items {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(item, item))
	}

	sort.Slice(buttons, func(i, j int) bool {
		return buttons[i].Text < buttons[j].Text
	})

	return buttons
}

// createKeyboard creates an inline keyboard with a given number of buttons per row.
func createKeyboard(
	_ context.Context, buttonsPerRow int, buttons []tgbotapi.InlineKeyboardButton, //nolint:unparam,nolintlint
) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < len(buttons); i += buttonsPerRow {
		end := i + buttonsPerRow
		if end > len(buttons) {
			end = len(buttons)
		}

		rows = append(rows, buttons[i:end])
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildMessageWithSubscription constructs a subscription message with optional labels.
func (h *BotHandler) buildMessageWithSubscription(subscription ds.SubscriptionResponse, isIncludeLabel bool) string {
	var sb strings.Builder

	sb.WriteString(h.formatSubscriptionField(
		subscription.Brand,
		"🚗",
		"Brand",
		isIncludeLabel),
	)
	sb.WriteString(h.formatSubscriptionField(
		strings.Join(subscription.Model, ", "),
		"🚘",
		"Models",
		isIncludeLabel),
	)
	sb.WriteString(h.formatSubscriptionField(
		fmt.Sprintf("%s€ - %s€", subscription.PriceFrom, subscription.PriceTo),
		"💰",
		"Price",
		isIncludeLabel),
	)
	sb.WriteString(h.formatSubscriptionField(
		fmt.Sprintf("%s - %s", subscription.YearFrom, subscription.YearTo),
		"📅",
		"Year",
		isIncludeLabel),
	)
	sb.WriteString(h.formatSubscriptionField(
		strings.Join(subscription.Region, ", "),
		"📍",
		"Regions",
		isIncludeLabel),
	)

	sb.WriteString("\n")

	return sb.String()
}

// formatSubscriptionField formats a subscription field with optional label.
func (h *BotHandler) formatSubscriptionField(value, emoji, label string, isIncludeLabel bool) string {
	if value == "" {
		return ""
	}

	if isIncludeLabel {
		return fmt.Sprintf("%s %s: %s, ", emoji, label, value)
	}

	return fmt.Sprintf("%s %s, ", emoji, value)
}
