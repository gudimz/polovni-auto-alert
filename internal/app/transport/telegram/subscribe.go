package telegram

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"

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
		LastMessageID   int
	}

	MessageWithButtonsParams struct {
		ChatID         int64
		Text           string
		ActionsButtons []tgbotapi.InlineKeyboardButton
		Buttons        []tgbotapi.InlineKeyboardButton
		ButtonsPerRow  int
		IsNeedEditMsg  bool
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

	brandButtonsPerRow     = 3
	modelButtonsPerRow     = 3
	chassisButtonsPerRow   = 3
	regionButtonsPerRow    = 3
	sendPriceButtonsPerRow = 2
	sendYearButtonsPerRow  = 2

	maxModelsPerPage = 36
	maxBrandsPerPage = 36
)

// handleSubscribe handles the /subscribe command, starting the subscription process.
func (h *BotHandler) handleSubscribe(ctx context.Context, chatID int64) error {
	return h.startSubscription(ctx, chatID)
}

func (h *BotHandler) startSubscription(ctx context.Context, chatID int64) error {
	h.state[chatID] = &SubscribeState{ //nolint:exhaustruct,nolintlint
		Step:            brandSelectionStep,
		InProgress:      true,
		SelectedModels:  []string{},
		SelectedChassis: []string{},
		SelectedRegions: []string{},
	}

	return h.sendBrandSelectionMessage(ctx, chatID, 0)
}

// sendBrandSelectionMessage sends a message asking the user to select a car brand.
func (h *BotHandler) sendBrandSelectionMessage(ctx context.Context, chatID int64, page int) error {
	text := `
🚗 Please choose a car brand:

You can cancel the process at any time by typing '🚫 cancel'.`

	brands := make([]string, 0, len(h.svc.GetCarsList()))
	for brand := range h.svc.GetCarsList() {
		brands = append(brands, brand)
	}

	sort.Slice(brands, func(i, j int) bool {
		return brands[i] < brands[j]
	})

	var keyboard tgbotapi.InlineKeyboardMarkup

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
	}

	// Check if pagination is needed
	if len(brands) > maxBrandsPerPage {
		totalPages := (len(brands) + maxBrandsPerPage - 1) / maxBrandsPerPage

		// Generate buttons for the current page
		start := page * maxModelsPerPage
		end := (page + 1) * maxModelsPerPage

		if end > len(brands) {
			end = len(brands)
		}

		buttons := generateButtonsFromSlice(ctx, brands[start:end])
		// Add pagination buttons
		paginationButtons := generatePaginationButtons(page, totalPages)

		keyboard = createKeyboardWithPagination(ctx, modelButtonsPerRow, actionsButtons, buttons, paginationButtons)
	} else {
		buttons := generateButtonsFromSlice(ctx, brands)
		keyboard = createKeyboard(ctx, brandButtonsPerRow, actionsButtons, buttons)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	// if it first times send message else edit this message
	isNeedEditMsg := false
	if _, exists := h.state[chatID]; exists {
		isNeedEditMsg = true
	}

	if err := h.sendSubscribeMessage(ctx, chatID, msg, isNeedEditMsg); err != nil {
		h.l.Error("failed to send brand selection message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send brand selection message")
	}

	return nil
}

// handleSelectBrand handles the brand selection step.
func (h *BotHandler) handleSelectBrand(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	data := callbackQuery.Data

	// Check if the callback is for pagination
	if strings.HasPrefix(data, "prev_page_") || strings.HasPrefix(data, "next_page_") {
		page, _ := strconv.Atoi(strings.Split(data, "_")[2])

		return h.sendBrandSelectionMessage(ctx, callbackQuery.Message.Chat.ID, page)
	}

	state := h.state[callbackQuery.From.ID]
	state.SelectedBrand = data
	state.Step = modelSelectionStep

	text := `
🚗 Please choose car models (you can select multiple). When you're done, type '✅ done':

You can cancel the process at any time by sending '🚫 cancel'`

	return h.sendModelSelectionMessage(ctx, callbackQuery.Message.Chat.ID, text, data, 0)
}

// sendModelSelectionMessage sends a message asking the user to select car models.
func (h *BotHandler) sendModelSelectionMessage(ctx context.Context, chatID int64, text, brand string, page int) error {
	var keyboard tgbotapi.InlineKeyboardMarkup

	models := h.svc.GetCarsList()[brand]

	sort.Slice(models, func(i, j int) bool {
		return models[i] < models[j]
	})

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "/done"),
	}

	// Check if pagination is needed
	if len(models) > maxModelsPerPage {
		totalPages := (len(models) + maxModelsPerPage - 1) / maxModelsPerPage

		// Generate buttons for the current page
		start := page * maxModelsPerPage
		end := (page + 1) * maxModelsPerPage

		if end > len(models) {
			end = len(models)
		}

		buttons := generateButtonsFromSlice(ctx, models[start:end])
		// Add pagination buttons
		paginationButtons := generatePaginationButtons(page, totalPages)

		keyboard = createKeyboardWithPagination(ctx, modelButtonsPerRow, actionsButtons, buttons, paginationButtons)
	} else {
		buttons := generateButtonsFromSlice(ctx, models)
		keyboard = createKeyboard(ctx, modelButtonsPerRow, actionsButtons, buttons)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if err := h.sendSubscribeMessage(ctx, chatID, msg, true); err != nil {
		h.l.Error("failed to send model selection message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send model selection message")
	}

	return nil
}

// generatePaginationButtons generates pagination buttons for navigating between pages.
func generatePaginationButtons(currentPage, totalPages int) []tgbotapi.InlineKeyboardButton {
	var paginationButtons []tgbotapi.InlineKeyboardButton

	if currentPage > 0 {
		paginationButtons = append(
			paginationButtons,
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Previous", fmt.Sprintf("prev_page_%d", currentPage-1)),
		)
	}

	if currentPage < totalPages-1 {
		paginationButtons = append(
			paginationButtons,
			tgbotapi.NewInlineKeyboardButtonData("➡️ Next", fmt.Sprintf("next_page_%d", currentPage+1)),
		)
	}

	return paginationButtons
}

// handleSelectModels handles the model selection step.
func (h *BotHandler) handleSelectModels(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	state := h.state[callbackQuery.From.ID]
	data := callbackQuery.Data

	isPagination := strings.HasPrefix(data, "prev_page_") || strings.HasPrefix(data, "next_page_")

	var text string

	if !isPagination { // Selecting a model
		if !contains(state.SelectedModels, data) {
			state.SelectedModels = append(state.SelectedModels, data)
		}

		text = fmt.Sprintf(`
🚗 Selected models: %s

Please choose more models or type '✅ done' if you are finished:
You can cancel the process at any time by typing '🚫 cancel'`,
			strings.Join(state.SelectedModels, ", "))
	} else { // Pagination
		text = callbackQuery.Message.Text
	}

	page := 0

	if isPagination {
		page, _ = strconv.Atoi(strings.Split(data, "_")[2])
	}

	return h.sendModelSelectionMessage(ctx, callbackQuery.Message.Chat.ID, text, state.SelectedBrand, page)
}

// sendChassisSelectionMessage sends a message asking the user to select a car chassis.
func (h *BotHandler) sendChassisSelectionMessage(ctx context.Context, chatID int64, text string) error {
	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "/done"),
	}

	buttons := generateButtons(ctx, h.svc.GetChassisList())

	if err := h.sendSubscribeMessageWithButtons(ctx, MessageWithButtonsParams{
		ChatID:         chatID,
		Text:           text,
		Buttons:        buttons,
		ActionsButtons: actionsButtons,
		ButtonsPerRow:  chassisButtonsPerRow,
		IsNeedEditMsg:  true,
	}); err != nil {
		h.l.Error("failed to send chassis selection message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send chassis selection message")
	}

	return nil
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

Please choose more chassis or type '✅ done' if you are finished:
You can cancel the process at any time by sending '🚫 cancel' or skip this step by typing '⏭️ skip'.`,
		strings.Join(state.SelectedChassis, ", "))

	return h.sendChassisSelectionMessage(ctx, callbackQuery.Message.Chat.ID, text)
}

// sendRegionSelectionMessage sends a message asking the user to select regions.
func (h *BotHandler) sendRegionSelectionMessage(ctx context.Context, chatID int64, text string) error {
	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "/done"),
	}

	buttons := generateButtons(ctx, h.svc.GetRegionsList())

	if err := h.sendSubscribeMessageWithButtons(ctx, MessageWithButtonsParams{
		ChatID:         chatID,
		Text:           text,
		Buttons:        buttons,
		ActionsButtons: actionsButtons,
		ButtonsPerRow:  chassisButtonsPerRow,
		IsNeedEditMsg:  false,
	}); err != nil {
		h.l.Error("failed to send region selection message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send region selection message")
	}

	return nil
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

Please choose more regions or type '✅ done' if you are finished:
You can cancel the process at any time by typing '🚫 cancel' or skip this step by sending '️️⏭️ skip'.`,
		strings.Join(state.SelectedRegions, ", "))

	return h.sendRegionSelectionMessage(ctx, callbackQuery.Message.Chat.ID, text)
}

// sendPriceFromMessage sends a message asking the user to enter the minimum price.
func (h *BotHandler) sendPriceFromMessage(ctx context.Context, chatID int64) error {
	text := `
💰 Please enter the minimum price in € or type '️⏭️ skip' to skip this step:

You can cancel the process at any time by sending '🚫 cancel'.`

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
	}

	if err := h.sendSubscribeMessageWithButtons(ctx, MessageWithButtonsParams{
		ChatID:         chatID,
		Text:           text,
		Buttons:        nil,
		ActionsButtons: actionsButtons,
		ButtonsPerRow:  chassisButtonsPerRow,
		IsNeedEditMsg:  false,
	}); err != nil {
		h.l.Error("failed to send price from message", logger.ErrAttr(err))
		return errors.New("failed to send price from message")
	}

	return nil
}

// handlePriceFrom processes the user's input for the minimum price.
func (h *BotHandler) handlePriceFrom(ctx context.Context, message *tgbotapi.Message) error {
	errText := "⚠️ Invalid price from. Please enter a valid number or type '️⏭️ skip' to skip this step:"
	if err := h.handleNumericInput(ctx, message, priceFromStep, priceToStep, errText, sendPriceButtonsPerRow); err != nil {
		return err
	}

	return h.sendPriceToMessage(ctx, message.Chat.ID)
}

// sendPriceToMessage sends a message asking the user to enter the maximum price.
func (h *BotHandler) sendPriceToMessage(ctx context.Context, chatID int64) error {
	text := `
💰 Please enter the maximum price in € or type '️⏭️ skip' to skip this step:

You can cancel the process at any time by sending '🚫 cancel'.`

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = createKeyboard(ctx, sendPriceButtonsPerRow, actionsButtons, nil)

	_, err := h.tgBot.SendMessage(msg)
	if err != nil {
		h.l.Error("failed to send price to message", logger.ErrAttr(err))
		return errors.New("failed to send price to message")
	}

	return nil
}

// handlePriceTo processes the user's input for the maximum price.
func (h *BotHandler) handlePriceTo(ctx context.Context, message *tgbotapi.Message) error {
	errText := "⚠️ Invalid price to. Please enter a valid number or type '️⏭️ skip' to skip this step:"
	if err := h.handleNumericInput(ctx, message, priceToStep, yearFromStep, errText, sendPriceButtonsPerRow); err != nil {
		return err
	}

	return h.sendYearFromMessage(ctx, message.Chat.ID)
}

// sendYearFromMessage sends a message asking the user to enter the start year.
func (h *BotHandler) sendYearFromMessage(ctx context.Context, chatID int64) error {
	text := `
📅 Please enter the start year or type '️⏭️ skip' to skip this step:

You can cancel the process at any time by sending '🚫 cancel'.`

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = createKeyboard(ctx, sendYearButtonsPerRow, actionsButtons, nil)

	if _, err := h.tgBot.SendMessage(msg); err != nil {
		h.l.Error("failed to send year from message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send year from message")
	}

	return nil
}

// handleYearFrom processes the user's input for the start year.
func (h *BotHandler) handleYearFrom(ctx context.Context, message *tgbotapi.Message) error {
	errText := "⚠️ Invalid year from. Please enter a valid number or type '️⏭️ skip' to skip this step:"
	if err := h.handleNumericInput(ctx, message, yearFromStep, yearToStep, errText, sendYearButtonsPerRow); err != nil {
		return err
	}

	return h.sendYearToMessage(ctx, message.Chat.ID)
}

// sendYearToMessage sends a message asking the user to enter the end year.
func (h *BotHandler) sendYearToMessage(ctx context.Context, chatID int64) error {
	text := `
📅 Please enter the end year or type '️⏭️ skip' to skip this step:

You can cancel the process at any time by sending '🚫 cancel'.`

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = createKeyboard(ctx, sendYearButtonsPerRow, actionsButtons, nil)

	if _, err := h.tgBot.SendMessage(msg); err != nil {
		h.l.Error("failed to send year to message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send year to message")
	}

	return nil
}

// handleYearTo processes the user's input for the end year.
func (h *BotHandler) handleYearTo(ctx context.Context, message *tgbotapi.Message) error {
	errText := "⚠️ Invalid year to. Please enter a valid number or type '️⏭️ skip' to skip this step:"
	if err := h.handleNumericInput(
		ctx, message, yearToStep, confirmSelectionStep, errText, sendYearButtonsPerRow,
	); err != nil {
		return err
	}

	return h.sendConfirmationMessage(ctx, message.Chat.ID)
}

// handleNumericInput handles the user's input for numeric values.
func (h *BotHandler) handleNumericInput(
	ctx context.Context,
	message *tgbotapi.Message,
	currStep subscribeStep,
	nextStep subscribeStep,
	errorMessage string,
	buttonsPerRow int, //nolint:unparam,nolintlint
) error {
	state := h.state[message.Chat.ID]
	input := message.Text

	if _, atoiErr := strconv.Atoi(input); atoiErr != nil {
		actionsButtons := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
			tgbotapi.NewInlineKeyboardButtonData("⏭️ Skip", "/skip"),
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, errorMessage)
		msg.ReplyMarkup = createKeyboard(ctx, buttonsPerRow, actionsButtons, nil)

		if _, err := h.tgBot.SendMessage(msg); err != nil {
			h.l.Error("failed to send validation message", logger.ErrAttr(err))
			return errors.Wrap(err, "failed to send validation message")
		}

		return atoiErr
	}

	switch currStep { //nolint:exhaustive,nolintlint
	case priceFromStep:
		state.PriceFrom = input
	case priceToStep:
		state.PriceTo = input
	case yearFromStep:
		state.YearFrom = input
	case yearToStep:
		state.YearTo = input
	}

	state.Step = nextStep

	return nil
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
		text := `
🚙 Please choose chassis (you can select multiple). When you're done, type '️✅ done':

You can cancel the process at any time by sending '🚫 cancel' or skip this step by sending '️⏭️ skip'.`

		return h.sendChassisSelectionMessage(ctx, chatID, text)
	case regionSelectionStep:
		text := `
📍 Please choose regions (you can select multiple). When you're done, type '️✅ done':

You can cancel the process at any time by sending '🚫 cancel' or skip this step by sending '️⏭️ skip'.`

		return h.sendRegionSelectionMessage(ctx, chatID, text)
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
		text := `
📍 Please choose regions (you can select multiple). When you're done, type '️✅ done':

You can cancel the process at any time by sending '🚫 cancel' or skip this step by sending '️⏭️ skip'.`

		return h.sendRegionSelectionMessage(ctx, chatID, text)
	case priceFromStep:
		// Clear the state for the previous step for the regions and move on to the next step
		state.SelectedRegions = state.SelectedRegions[:0]
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
func (h *BotHandler) sendConfirmationMessage(ctx context.Context, chatID int64) error {
	state := h.state[chatID]
	text := fmt.Sprintf(`
	🚗 Brand: %s
	🚘 Models: %s
	🚙 Chassis: %s
	📍 Regions: %s
	💰 Price: %s€ - %s€
	📅 Year: %s - %s

	Please type '✅ confirm' to save this subscription or '🚫 cancel' to discard it.`,
		state.SelectedBrand,
		strings.Join(state.SelectedModels, ", "),
		strings.Join(state.SelectedChassis, ", "),
		strings.Join(state.SelectedRegions, ", "),
		state.PriceFrom, state.PriceTo,
		state.YearFrom, state.YearTo,
	)

	actionsButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🚫 Cancel", "/cancel"),
		tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "/confirm"),
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = createKeyboard(ctx, sendYearButtonsPerRow, actionsButtons, nil)

	if _, err := h.tgBot.SendMessage(msg); err != nil {
		h.l.Error("failed to send confirm message", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to send confirm message")
	}

	return nil
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

// sendSubscribeMessage sends a message for subscribe new listings.
func (h *BotHandler) sendSubscribeMessage(
	_ context.Context,
	chatID int64,
	msg tgbotapi.MessageConfig,
	isNeedEditMsg bool,
) error {
	state := h.state[chatID]

	if state != nil && state.LastMessageID != 0 && isNeedEditMsg {
		replyMarkup, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		if !ok {
			return errors.New("failed to cast reply markup to InlineKeyboardMarkup")
		}

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, state.LastMessageID, msg.Text, replyMarkup)

		_, err := h.tgBot.SendMessage(editMsg)
		if err != nil {
			return errors.Wrap(err, "failed to edit message")
		}
	} else {
		sendMsg, err := h.tgBot.SendMessage(msg)
		if err != nil {
			return errors.Wrap(err, "failed to send brand selection message")
		}

		h.state[chatID].LastMessageID = sendMsg.MessageID
	}

	return nil
}

func (h *BotHandler) sendSubscribeMessageWithButtons(ctx context.Context, params MessageWithButtonsParams) error {
	keyboard := createKeyboard(ctx, params.ButtonsPerRow, params.ActionsButtons, params.Buttons)
	msg := tgbotapi.NewMessage(params.ChatID, params.Text)
	msg.ReplyMarkup = keyboard

	return h.sendSubscribeMessage(ctx, params.ChatID, msg, params.IsNeedEditMsg)
}
