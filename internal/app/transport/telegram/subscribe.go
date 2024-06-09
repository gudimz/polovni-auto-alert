package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type (
	subscribeStep int

	SubscribeState struct {
		Step            subscribeStep
		InProgress      bool
		SelectedBrand   string
		SelectedModels  []string
		SelectedChassis []string
		SelectedRegions []string
		PriceFrom       string
		PriceTo         string
		YearFrom        string
		YearTo          string
	}
)

const (
	brandSelectionStep   subscribeStep = 1
	modelSelectionStep   subscribeStep = 2
	chassisSelectionStep subscribeStep = 3
	regionSelectionStep  subscribeStep = 4
	priceFromStep        subscribeStep = 5
	priceToStep          subscribeStep = 6
	yearFromStep         subscribeStep = 7
	yearToStep           subscribeStep = 8
	confirmSelectionStep subscribeStep = 9

	brandButtonsPerRow   = 3
	modelButtonsPerRow   = 3
	chassisButtonsPerRow = 3
	regionButtonsPerRow  = 3
)

// handleSubscribe handles the /subscribe command, starting the subscription process.
func (h *BotHandler) handleSubscribe(ctx context.Context, chatID int64) error {
	return h.startSubscription(ctx, chatID)
}

func (h *BotHandler) startSubscription(ctx context.Context, chatID int64) error {
	h.state[chatID] = &SubscribeState{
		Step:            brandSelectionStep,
		InProgress:      true,
		SelectedModels:  []string{},
		SelectedChassis: []string{},
		SelectedRegions: []string{},
	}
	return h.sendBrandSelectionMessage(ctx, chatID)
}

// sendBrandSelectionMessage sends a message asking the user to select a car brand.
func (h *BotHandler) sendBrandSelectionMessage(ctx context.Context, chatID int64) error {
	text := `
🚗 Please choose a car brand:

You can cancel the process at any time by sending /cancel`
	brands := make([]string, 0, len(h.svc.GetCarsList()))
	for brand := range h.svc.GetCarsList() {
		brands = append(brands, brand)
	}

	buttons := generateButtonsFromSlice(ctx, brands)
	keyboard := createKeyboard(ctx, brandButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send brand selection message", logger.ErrAttr(err))
	}

	return err
}

// handleSelectBrand handles the brand selection step.
func (h *BotHandler) handleSelectBrand(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	brand := callbackQuery.Data
	state := h.state[callbackQuery.From.ID]
	state.SelectedBrand = brand
	state.Step = modelSelectionStep

	return h.sendModelSelectionMessage(ctx, callbackQuery.Message.Chat.ID, brand)
}

// sendModelSelectionMessage sends a message asking the user to select car models.
func (h *BotHandler) sendModelSelectionMessage(ctx context.Context, chatID int64, brand string) error {
	text := `
🚗 Please choose car models (you can select multiple). When you're done, type /done:

You can cancel the process at any time by sending /cancel`

	models := h.svc.GetCarsList()[brand]
	buttons := generateButtonsFromSlice(ctx, models)
	keyboard := createKeyboard(ctx, modelButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send model selection message", logger.ErrAttr(err))
	}

	return err
}

// handleSelectModels handles the model selection step.
func (h *BotHandler) handleSelectModels(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	state := h.state[callbackQuery.From.ID]
	model := callbackQuery.Data

	if !contains(state.SelectedModels, model) {
		state.SelectedModels = append(state.SelectedModels, model)
	}

	text := fmt.Sprintf(`
🚗 Selected models: %s

Please choose more models or type /done if you are finished:
You can cancel the process at any time by sending /cancel`,
		strings.Join(state.SelectedModels, ", "))

	buttons := generateButtonsFromSlice(ctx, h.svc.GetCarsList()[state.SelectedBrand])
	keyboard := createKeyboard(ctx, modelButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send updated model selection message", logger.ErrAttr(err))
	}

	return err
}

// sendChassisSelectionMessage sends a message asking the user to select a car chassis.
func (h *BotHandler) sendChassisSelectionMessage(ctx context.Context, chatID int64) error {
	text := `
🚙 Please choose chassis (you can select multiple). When you're done, type /done:

You can cancel the process at any time by sending /cancel or skip this step by sending /skip.`

	buttons := generateButtons(ctx, h.svc.GetChassisList())
	keyboard := createKeyboard(ctx, chassisButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send chassis selection message", logger.ErrAttr(err))
	}

	return err
}

// handleSelectChassis handles the chassis selection step.
func (h *BotHandler) handleSelectChassis(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	state := h.state[callbackQuery.From.ID]
	chassis := callbackQuery.Data

	if !contains(state.SelectedChassis, chassis) {
		state.SelectedChassis = append(state.SelectedChassis, chassis)
	}

	text := fmt.Sprintf(`
🚙 Selected chassis: %s

Please choose more chassis or type /done if you are finished:
You can cancel the process at any time by sending /cancel or skip this step by sending /skip.`,
		strings.Join(state.SelectedChassis, ", "))

	buttons := generateButtons(ctx, h.svc.GetChassisList())
	keyboard := createKeyboard(ctx, chassisButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send updated chassis selection message", logger.ErrAttr(err))
	}

	return err
}

// sendRegionSelectionMessage sends a message asking the user to select regions.
func (h *BotHandler) sendRegionSelectionMessage(ctx context.Context, chatID int64) error {
	text := `
📍 Please choose regions (you can select multiple). When you're done, type /done:

You can cancel the process at any time by sending /cancel or skip this step by sending /skip.`

	buttons := generateButtons(ctx, h.svc.GetRegionsList())
	keyboard := createKeyboard(ctx, regionButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send region selection message", logger.ErrAttr(err))
	}

	return err
}

// handleSelectRegions handles the region selection step.
func (h *BotHandler) handleSelectRegions(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	state := h.state[callbackQuery.From.ID]
	region := callbackQuery.Data

	if !contains(state.SelectedRegions, region) {
		state.SelectedRegions = append(state.SelectedRegions, region)
	}

	text := fmt.Sprintf(`
📍 Selected regions: %s

Please choose more regions or type /done if you are finished:
You can cancel the process at any time by sending /cancel or skip this step by sending /skip.`,
		strings.Join(state.SelectedRegions, ", "))

	buttons := generateButtons(ctx, h.svc.GetRegionsList())
	keyboard := createKeyboard(ctx, regionButtonsPerRow, buttons)

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send updated region selection message", logger.ErrAttr(err))
	}

	return err
}

// sendPriceFromMessage sends a message asking the user to enter the minimum price.
func (h *BotHandler) sendPriceFromMessage(_ context.Context, chatID int64) error {
	text := `
💰 Please enter the minimum price in € or type /skip to skip this step:

You can cancel the process at any time by sending /cancel.`

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send price from message", logger.ErrAttr(err))
	}
	return err
}

// handlePriceFrom processes the user's input for the minimum price.
func (h *BotHandler) handlePriceFrom(ctx context.Context, message *tgbotapi.Message) error {
	state := h.state[message.Chat.ID]
	priceFrom := message.Text

	if _, err := strconv.Atoi(priceFrom); err != nil {
		text := "⚠️ Invalid price. Please enter a valid number or type /skip to skip this step:"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		_, err = h.tgBot.SendMessage(msg)
		if err != nil {
			h.l.Error("failed to send price from validation message", logger.ErrAttr(err))
		}

		return err
	}

	state.PriceFrom = priceFrom
	state.Step = priceToStep
	return h.sendPriceToMessage(ctx, message.Chat.ID)
}

// sendPriceToMessage sends a message asking the user to enter the maximum price.
func (h *BotHandler) sendPriceToMessage(_ context.Context, chatID int64) error {
	text := `
💰 Please enter the maximum price in € or type /skip to skip this step:

You can cancel the process at any time by sending /cancel.`

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send price to message", logger.ErrAttr(err))
	}
	return err
}

// handlePriceTo processes the user's input for the maximum price.
func (h *BotHandler) handlePriceTo(ctx context.Context, message *tgbotapi.Message) error {
	state := h.state[message.Chat.ID]
	priceTo := message.Text

	if _, err := strconv.Atoi(priceTo); err != nil {
		text := "⚠️ Invalid price. Please enter a valid number or type /skip to skip this step:"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		_, err = h.tgBot.SendMessage(msg)
		if err != nil {
			h.l.Error("failed to send price to validation message", logger.ErrAttr(err))
		}
		return err
	}

	state.PriceTo = priceTo
	state.Step = yearFromStep
	return h.sendYearFromMessage(ctx, message.Chat.ID)
}

// sendYearFromMessage sends a message asking the user to enter the start year.
func (h *BotHandler) sendYearFromMessage(_ context.Context, chatID int64) error {
	text := `
📅 Please enter the start year or type /skip to skip this step:

You can cancel the process at any time by sending /cancel.`

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send year from message", logger.ErrAttr(err))
	}
	return err
}

// handleYearFrom processes the user's input for the start year.
func (h *BotHandler) handleYearFrom(ctx context.Context, message *tgbotapi.Message) error {
	state := h.state[message.Chat.ID]
	yearFrom := message.Text

	if _, err := strconv.Atoi(yearFrom); err != nil {
		text := "⚠️ Invalid year. Please enter a valid number or type /skip to skip this step:"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		_, err = h.tgBot.SendMessage(msg)
		if err != nil {
			h.l.Error("failed to send year from validation message", logger.ErrAttr(err))
		}
		return err
	}

	state.YearFrom = yearFrom
	state.Step = yearToStep
	return h.sendYearToMessage(ctx, message.Chat.ID)
}

// sendYearToMessage sends a message asking the user to enter the end year.
func (h *BotHandler) sendYearToMessage(_ context.Context, chatID int64) error {
	text := `
📅 Please enter the end year or type /skip to skip this step:

You can cancel the process at any time by sending /cancel.`

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send year to message", logger.ErrAttr(err))
	}
	return err
}

// handleYearTo processes the user's input for the end year.
func (h *BotHandler) handleYearTo(ctx context.Context, message *tgbotapi.Message) error {
	state := h.state[message.Chat.ID]
	yearTo := message.Text

	if _, err := strconv.Atoi(yearTo); err != nil {
		text := "⚠️ Invalid year. Please enter a valid number or type /skip to skip this step:"
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		_, err = h.tgBot.SendMessage(msg)
		if err != nil {
			h.l.Error("failed to send year to validation message", logger.ErrAttr(err))
		}
		return err
	}

	state.YearTo = yearTo
	state.Step = confirmSelectionStep
	return h.sendConfirmationMessage(ctx, message.Chat.ID)
}

// handleCancel handles the /cancel command, canceling the subscription process.
func (h *BotHandler) handleCancel(_ context.Context, chatID int64) error {
	state, exists := h.state[chatID]
	if !exists || !state.InProgress {
		return nil // Ignore if subscription process is not in progress
	}

	delete(h.state, chatID)

	text := "🚫 Subscription process has been cancelled."

	return h.sendMessage(chatID, text, handleNameCancel)
}

// handleDone handles the /done command, moving to the next step.
func (h *BotHandler) handleDone(ctx context.Context, chatID int64) error {
	state, exists := h.state[chatID]
	if !exists {
		return h.sendUnknownCommandMessage(ctx, chatID)
	}

	if !state.InProgress {
		return h.sendUnknownCommandMessage(ctx, chatID)
	}

	state.Step++

	switch state.Step {
	case brandSelectionStep:
		return nil
	case modelSelectionStep:
		return nil
	case chassisSelectionStep:
		return h.sendChassisSelectionMessage(ctx, chatID)
	case regionSelectionStep:
		return h.sendRegionSelectionMessage(ctx, chatID)
	case priceFromStep:
		return h.sendPriceFromMessage(ctx, chatID)
	case priceToStep:
		return h.sendPriceToMessage(ctx, chatID)
	case yearFromStep:
		return h.sendYearFromMessage(ctx, chatID)
	case yearToStep:
		return h.sendYearToMessage(ctx, chatID)
	case confirmSelectionStep:
		return h.sendConfirmationMessage(ctx, chatID)
	default:
		h.l.Warn("unknown subscription step", logger.AnyAttr("step", state.Step))
		text := "⚠️ An unknown error occurred. The subscription process has been cancelled. Please try again."
		if err := h.sendMessage(chatID, text, handleNameDone); err != nil {
			h.l.Error("failed to send cancellation message", logger.ErrAttr(err))
		}

		delete(h.state, chatID)

		return nil
	}
}

// handleSkip handles the /skip command, allowing the user to skip optional steps.
func (h *BotHandler) handleSkip(ctx context.Context, chatID int64) error {
	state, exists := h.state[chatID]
	if !exists || !state.InProgress {
		return nil // Ignore if subscription process is not in progress
	}

	state.Step++
	switch state.Step { //nolint:exhaustive,nolintlint
	case regionSelectionStep:
		// Clear the state for the previous step for the chassis and move on to the next step
		state.SelectedChassis = state.SelectedChassis[:0]
		return h.sendRegionSelectionMessage(ctx, chatID)
	case priceFromStep:
		return h.sendPriceFromMessage(ctx, chatID)
	case priceToStep:
		return h.sendPriceToMessage(ctx, chatID)
	case yearFromStep:
		return h.sendYearFromMessage(ctx, chatID)
	case yearToStep:
		return h.sendYearToMessage(ctx, chatID)
	case confirmSelectionStep:
		return h.sendConfirmationMessage(ctx, chatID)
	default:
		h.l.Warn("unknown subscription step", logger.AnyAttr("step", state.Step))
	}

	return nil
}

// sendConfirmationMessage sends a message asking the user to confirm their subscription.
func (h *BotHandler) sendConfirmationMessage(_ context.Context, chatID int64) error {
	state := h.state[chatID]
	text := fmt.Sprintf(`
	🚗 Brand: %s
	🚘 Models: %s
	🚙 Chassis: %s
	📍 Regions: %s
	💰 Price: %s€ - %s€
	📅 Year: %s - %s

	Please type /confirm to save this subscription or /cancel to discard it.
	`,
		state.SelectedBrand,
		strings.Join(state.SelectedModels, ", "),
		strings.Join(state.SelectedChassis, ", "),
		strings.Join(state.SelectedRegions, ", "),
		state.PriceFrom, state.PriceTo,
		state.YearFrom, state.YearTo,
	)

	return h.sendMessage(chatID, text, handleNameConfirm)
}

// handleConfirm handles the /confirm command, saving the subscription to the database.
func (h *BotHandler) handleConfirm(ctx context.Context, chatID int64) error {
	state, exists := h.state[chatID]
	if !exists {
		return h.sendUnknownCommandMessage(ctx, chatID)
	}

	subscription := ds.SubscriptionRequest{
		UserID:    chatID,
		Brand:     state.SelectedBrand,
		Model:     state.SelectedModels,
		Chassis:   state.SelectedChassis,
		Region:    state.SelectedRegions,
		PriceFrom: state.PriceFrom,
		PriceTo:   state.PriceTo,
		YearFrom:  state.YearFrom,
		YearTo:    state.YearTo,
	}

	_, err := h.svc.CreateSubscription(ctx, subscription)
	if err != nil {
		text := "⚠️ An internal error occurred while saving your subscription. Please try again later."
		return h.sendMessage(chatID, text, handleNameConfirm)
	}

	delete(h.state, chatID)

	text := "✅ Your subscription has been saved successfully!"
	return h.sendMessage(chatID, text, handleNameConfirm)
}
