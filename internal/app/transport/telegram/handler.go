package telegram

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type BotHandler struct {
	l     *logger.Logger
	tgBot TgBot
	svc   Service
	state map[int64]*SubscribeState
}

const (
	handleNameStart             = "/start"
	handleNameStop              = "/stop"
	handleNameListSubscriptions = "/list_subscriptions"
	handleNameSubscribe         = "/subscribe"
	handleNameUnsubscribe       = "/unsubscribe"
	handleNameCancel            = "/cancel"
	handleNameSkip              = "/skip"
	handleNameDone              = "/done"
	handleNameConfirm           = "/confirm"
	handleNameUnknown           = "unknown command"
)

func NewBotHandler(l *logger.Logger, tgBot TgBot, svc Service) *BotHandler {
	return &BotHandler{
		l:     l,
		tgBot: tgBot,
		svc:   svc,
		state: make(map[int64]*SubscribeState),
	}
}

// Start begins processing updates from Telegram.
func (h *BotHandler) Start(ctx context.Context) error {
	h.l.Info("bot handler started")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = h.tgBot.GetCfg().UpdateCfgTimeout

	updates := h.tgBot.GetAPI().GetUpdatesChan(updateConfig)

	for {
		select {
		case update := <-updates:
			h.HandleUpdate(ctx, update)
		case <-ctx.Done():
			h.l.Info("bot handler stopped", logger.ErrAttr(ctx.Err()))
			return nil
		}
	}
}

// HandleUpdate processes incoming updates from Telegram.
func (h *BotHandler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer h.recoverPanic()

	switch {
	case update.Message != nil:
		h.handleMessage(ctx, update.Message)
	case update.CallbackQuery != nil:
		h.handleCallbackQuery(ctx, update.CallbackQuery)
	}
}

// handleMessage processes incoming text messages.
func (h *BotHandler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	var err error
	switch message.Text {
	case handleNameStart:
		err = h.handleStart(ctx, message.Chat)
	case handleNameStop:
		err = h.handleStop(ctx, message.Chat.ID)
	case handleNameSubscribe:
		err = h.handleSubscribe(ctx, message.Chat.ID)
	case handleNameUnsubscribe:
		err = h.handleUnsubscribe(ctx, message.Chat.ID)
	case handleNameListSubscriptions:
		err = h.handleListSubscriptions(ctx, message.Chat.ID)
	case handleNameDone:
		err = h.handleDone(ctx, message.Chat.ID)
	case handleNameCancel:
		err = h.handleCancel(ctx, message.Chat.ID)
	case handleNameSkip:
		err = h.handleSkip(ctx, message.Chat.ID)
	case handleNameConfirm:
		err = h.handleConfirm(ctx, message.Chat.ID)
	default:
		state, exists := h.state[message.Chat.ID]
		if exists {
			switch state.Step { //nolint:exhaustive,nolintlint
			case priceFromStep:
				err = h.handlePriceFrom(ctx, message)
			case priceToStep:
				err = h.handlePriceTo(ctx, message)
			case yearFromStep:
				err = h.handleYearFrom(ctx, message)
			case yearToStep:
				err = h.handleYearTo(ctx, message)
			default:
				err = h.sendUnknownCommandMessage(ctx, message.Chat.ID)
			}
		} else {
			err = h.sendUnknownCommandMessage(ctx, message.Chat.ID)
		}
	}

	if err != nil {
		h.l.Error("failed to send message", logger.ErrAttr(err))
	}
}

func (h *BotHandler) handleCallbackQuery(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) {
	if strings.HasPrefix(callbackQuery.Data, handleNameUnsubscribe+":") {
		callbackQuery.Data = strings.TrimPrefix(callbackQuery.Data, handleNameUnsubscribe+":")
		if err := h.handleUnsubscribeCallback(ctx, callbackQuery); err != nil {
			h.l.Error("failed to handle unsubscribe callback", logger.ErrAttr(err))
		}
		return
	}
	state, exists := h.state[callbackQuery.From.ID]
	if !exists {
		if err := h.sendUnknownCommandMessage(ctx, callbackQuery.Message.Chat.ID); err != nil {
			h.l.Error("failed to send unknown command message", logger.ErrAttr(err))
		}
		return
	}

	var err error
	switch state.Step { //nolint:exhaustive,nolintlint
	case brandSelectionStep:
		err = h.handleSelectBrand(ctx, callbackQuery)
	case modelSelectionStep:
		err = h.handleSelectModels(ctx, callbackQuery)
	case chassisSelectionStep:
		err = h.handleSelectChassis(ctx, callbackQuery)
	case regionSelectionStep:
		err = h.handleSelectRegions(ctx, callbackQuery)
	case priceFromStep:
		err = h.handlePriceFrom(ctx, callbackQuery.Message)
	case priceToStep:
		err = h.handlePriceTo(ctx, callbackQuery.Message)
	case yearFromStep:
		err = h.handleYearFrom(ctx, callbackQuery.Message)
	case yearToStep:
		err = h.handleYearTo(ctx, callbackQuery.Message)
	default:
		h.l.Warn("unknown subscription step", logger.AnyAttr("step", state.Step))
	}

	if err != nil {
		h.l.Error("failed to send message", logger.ErrAttr(err))
	}
}

// sendUnknownCommandMessage sends a message indicating an unknown command.
func (h *BotHandler) sendUnknownCommandMessage(_ context.Context, chatID int64) error {
	text := "🤔 I'm not sure what you mean. Please use one of the available commands."
	return h.sendMessage(chatID, text, handleNameUnknown)
}

// sendMessage sends a text message to the user.
func (h *BotHandler) sendMessage(chatID int64, text, handlerName string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error(handlerName+": failed to send message", logger.ErrAttr(err))
	}

	return err
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

// recoverPanic recovers from a panic and logs the error.
func (h *BotHandler) recoverPanic() {
	if r := recover(); r != nil {
		h.l.Error(
			"Recovered from panic",
			logger.AnyAttr("error", r),
			logger.StringAttr("stacktrace", string(debug.Stack())),
		)
	}
}
