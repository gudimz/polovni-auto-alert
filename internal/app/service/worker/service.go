package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/guregu/null"
	pkgerrors "github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

// Service represents the worker service which processes listings periodically.
type Service struct {
	l        *logger.Logger
	repo     Repository
	tgBot    TgBot
	interval time.Duration
}

var errBotBlockedByUser = pkgerrors.New("bot is blocked by user")

// NewService creates a new Worker Service instance.
func NewService(l *logger.Logger, repo Repository, tgBot TgBot, interval time.Duration) *Service {
	return &Service{
		l:        l,
		repo:     repo,
		tgBot:    tgBot,
		interval: interval,
	}
}

// Start begins the worker process, which periodically processes listings.
func (s *Service) Start(ctx context.Context) error {
	s.l.Info("worker interval set to", logger.DurationAttr("interval", s.interval))

	ticker := time.NewTicker(s.interval)

	go func() {
		defer s.recoverPanic()

		for {
			select {
			case <-ticker.C:
				s.l.Info("worker ticker ticked")

				if err := s.ProcessListings(ctx); err != nil {
					s.l.Error("failed to process listings", logger.ErrAttr(err))
				}
			case <-ctx.Done():
				ticker.Stop()
				s.l.Info("worker stopped", logger.ErrAttr(ctx.Err()))

				return
			}
		}
	}()

	s.l.Info("worker service started")

	return nil
}

// ProcessListings processes listings that need to be sent, notifies via Telegram,
// and updates their status in the repository.
func (s *Service) ProcessListings(ctx context.Context) error {
	listings, err := s.repo.GetListingsByIsNeedSend(ctx, true)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to get listings")
	}

	for _, listing := range listings {
		var subscription ds.SubscriptionResponse

		subscription, err = s.repo.GetSubscriptionByID(ctx, listing.SubscriptionID)
		if err != nil {
			s.l.Error("failed to get subscription",
				logger.ErrAttr(err),
				logger.StringAttr("subscription_id", listing.SubscriptionID),
			)

			continue
		}

		notification := ds.CreateNotificationRequest{
			SubscriptionID: listing.SubscriptionID,
			ListingID:      listing.ListingID,
			Status:         ds.StatusSent,
			Reason:         "",
		}

		if err = s.sendListing(ctx, subscription.UserID, listing); err != nil {
			s.l.Error("failed to send listing",
				logger.ErrAttr(err),
				logger.Int64Attr("user_id", subscription.UserID),
				logger.StringAttr("listing_id", listing.ListingID),
			)

			notification.Status = ds.StatusFailed
			notification.Reason = err.Error()
		}

		if errors.Is(err, errBotBlockedByUser) {
			continue
		}

		_, err = s.repo.CreateNotification(ctx, notification)
		if err != nil {
			s.l.Error("failed to create notification",
				logger.ErrAttr(err),
				logger.Int64Attr("user_id", subscription.UserID),
				logger.StringAttr("listing_id", listing.ListingID),
			)
		}

		isNeedSend := false
		newPrice := null.NewString("", false)
		// if the notification failed, we need to send it again
		if notification.Status == ds.StatusFailed {
			isNeedSend = true
			newPrice = listing.NewPrice
		}

		if err = s.repo.UpsertListing(ctx, ds.UpsertListingRequest{
			ListingID:      listing.ListingID,
			SubscriptionID: listing.SubscriptionID,
			Title:          listing.Title,
			Price:          listing.Price,
			NewPrice:       newPrice,
			EngineVolume:   listing.EngineVolume,
			Transmission:   listing.Transmission,
			BodyType:       listing.BodyType,
			Mileage:        listing.Mileage,
			Location:       listing.Location,
			Link:           listing.Link,
			Date:           listing.Date,
			IsNeedSend:     isNeedSend,
		}); err != nil {
			s.l.Error("failed to update listing",
				logger.ErrAttr(err),
				logger.Int64Attr("user_id", subscription.UserID),
				logger.StringAttr("listing_id", listing.ListingID),
				logger.StringAttr("subscription_id", listing.SubscriptionID),
			)
		}
	}

	s.l.Info("processed all listings")

	return nil
}

// sendListing sends a listing message to the user's tg with all the details.
func (s *Service) sendListing(ctx context.Context, chatID int64, listing ds.ListingResponse) error {
	price := listing.Price

	if listing.NewPrice.Valid && listing.NewPrice.String != listing.Price {
		direction := "🔺"
		if listing.NewPrice.String < listing.Price {
			direction = "🔻"
		}

		price = fmt.Sprintf("⚠%s%s%s", listing.Price, direction, listing.NewPrice.String)
	}

	text := fmt.Sprintf(`

	%s

	📝 *Title:* %s
	💰 *Price:* %s
	🏎️ *Engine Volume:* %s
	⚙️ *Transmission:* %s
	🚗 *Body Type:* %s
	🧭 *Mileage:* %s
	📍 *Location:* %s
	📅 *Date:* %s
	🌐 *Link:* [tap to link](%s)
	`,
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "👋 Hi, here's a new listing for your subscription."),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.Title),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, price),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.EngineVolume),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.Transmission),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.BodyType),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.Mileage),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.Location),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.Date.Format(time.DateTime)),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, listing.Link),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := s.tgBot.SendMessage(msg)
	if err != nil {
		var apiErr *tgbotapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusForbidden { // user blocked tg bot
			if err = s.RemoveAllSubscriptionsByUserID(ctx, chatID); err != nil {
				s.l.Warn("failed to remove all subscriptions by userID",
					logger.ErrAttr(err),
					logger.Int64Attr("user_id", chatID),
				)

				return errors.Join(errBotBlockedByUser, err)
			}

			return errBotBlockedByUser
		}

		return err
	}

	return nil
}

// RemoveAllSubscriptionsByUserID removes all subscriptions and associated listings for a given user.
func (s *Service) RemoveAllSubscriptionsByUserID(ctx context.Context, userID int64) error {
	subscriptions, err := s.repo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to get subscriptions by user id")
	}

	ids := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		ids[i] = sub.ID
	}

	if err = s.repo.DeleteListingsBySubscriptionIDs(ctx, ids); err != nil {
		return pkgerrors.Wrap(err, "failed to delete listings by subscription ids")
	}

	if err = s.repo.DeleteSubscriptionsByUserID(ctx, userID); err != nil {
		return pkgerrors.Wrap(err, "failed to delete subscriptions by user id")
	}

	if err = s.repo.DeleteUserByID(ctx, userID); err != nil {
		return pkgerrors.Wrap(err, "failed to delete user by id")
	}

	// TODO add transactions in future

	return nil
}

// recoverPanic recovers from a panic and logs the error.
func (s *Service) recoverPanic() {
	if r := recover(); r != nil {
		s.l.Error(
			"Recovered from panic",
			logger.AnyAttr("error", r),
			logger.StringAttr("stacktrace", string(debug.Stack())),
		)
	}
}
