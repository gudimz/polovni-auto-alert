package telegram

import (
	"context"
	"runtime/debug"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
