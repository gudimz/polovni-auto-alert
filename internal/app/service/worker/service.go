package worker

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"

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
		return errors.Wrap(err, "failed to get listings")
	}

	for _, listing := range listings {
		var subscription ds.SubscriptionResponse

		subscription, err = s.repo.GetSubscriptionByID(ctx, listing.SubscriptionID)
		if err != nil {
			s.l.Error("failed to get subscription",
				logger.ErrAttr(err),
				logger.StringAttr("subscription_id", listing.SubscriptionID),
			)
		}

		notification := ds.CreateNotificationRequest{
			ListingID: listing.ID,
			Status:    ds.StatusSent,
			Reason:    "",
		}

		if err = s.sendListing(ctx, subscription.UserID, listing); err != nil {
			s.l.Error("failed to send listing",
				logger.ErrAttr(err),
				logger.Int64Attr("user_id", subscription.UserID),
				logger.StringAttr("listing_id", listing.ID),
			)

			notification.Status = ds.StatusFailed
			notification.Reason = err.Error()
		}

		_, err = s.repo.CreateNotification(ctx, notification)
		if err != nil {
			s.l.Error("failed to create notification",
				logger.ErrAttr(err),
				logger.Int64Attr("user_id", subscription.UserID),
				logger.StringAttr("listing_id", listing.ID),
			)
		}

		if err = s.repo.UpsertListing(ctx, ds.UpsertListingRequest{
			ID:             listing.ID,
			SubscriptionID: listing.SubscriptionID,
			Title:          listing.Title,
			Price:          listing.Price,
			EngineVolume:   listing.EngineVolume,
			Transmission:   listing.Transmission,
			BodyType:       listing.BodyType,
			Mileage:        listing.Mileage,
			Location:       listing.Location,
			Link:           listing.Link,
			Date:           listing.Date,
			IsNeedSend:     false,
		}); err != nil {
			s.l.Error("failed to update listing",
				logger.ErrAttr(err),
				logger.Int64Attr("user_id", subscription.UserID),
				logger.StringAttr("listing_id", listing.ID),
				logger.StringAttr("subscription_id", listing.SubscriptionID),
			)
		}
	}

	s.l.Info("processed all listings")

	return nil
}

// sendListing sends a listing message to the user's tg with all the details.
func (s *Service) sendListing(ctx context.Context, chatID int64, listing ds.ListingResponse) error {
	text := fmt.Sprintf(`

	👋 Hi, here's a new listing for your subscription.

	📝 Title: %s
	💰 Price: %s€
	🏎️ Engine Volume: %s
	⚙️ Transmission: %s
	🚗 Body Type: %s
	🧭 Mileage: %s
	📍 Location: %s
	📅 Date: %s
	🌐 Link: %s
	`,
		listing.Title,
		listing.Price,
		listing.EngineVolume,
		listing.Transmission,
		listing.BodyType,
		listing.Mileage,
		listing.Location,
		listing.Date.Format(time.DateTime),
		listing.Link,
	)

	msg := tgbotapi.NewMessage(chatID, text)

	_, err := s.tgBot.SendMessage(msg)
	if err != nil {
		var apiErr *tgbotapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusForbidden { // user blocked tg bot
			if err = s.RemoveAllSubscriptionsByUserID(ctx, chatID); err != nil {
				s.l.Warn("failed to remove all subscriptions by userID",
					logger.ErrAttr(err),
					logger.Int64Attr("user_id", chatID),
				)
			}
		}

		return err
	}

	return nil
}

// RemoveAllSubscriptionsByUserID removes all subscriptions and associated listings for a given user.
func (s *Service) RemoveAllSubscriptionsByUserID(ctx context.Context, userID int64) error {
	subscriptions, err := s.repo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get subscriptions by user id")
	}

	ids := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		ids[i] = sub.ID
	}

	if err = s.repo.DeleteListingsBySubscriptionIDs(ctx, ids); err != nil {
		return errors.Wrap(err, "failed to delete listings by subscription ids")
	}

	if err = s.repo.DeleteSubscriptionsByUserID(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to delete subscriptions by user id")
	}

	if err = s.repo.DeleteUserByID(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to delete user by id")
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
