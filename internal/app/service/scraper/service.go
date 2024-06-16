package scraper

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

type Service struct {
	l         *logger.Logger
	repo      Repository
	paAdapter PolovniAutoAdapter
	interval  time.Duration
}

// NewService creates a new Scraper Service instance.
func NewService(l *logger.Logger, repo Repository, paAdapter PolovniAutoAdapter, interval time.Duration) *Service {
	return &Service{
		l:         l,
		repo:      repo,
		paAdapter: paAdapter,
		interval:  interval,
	}
}

// Start begins the scraping process.
func (s *Service) Start(ctx context.Context) error {
	s.l.Info("scraper interval set to", logger.DurationAttr("interval", s.interval))

	// To distinguish new listings from old ones,
	// we save all listings in the database when we start the app
	err := s.ScrapeAllListings(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(s.interval)

	go func() {
		defer s.recoverPanic()

		for {
			select {
			case <-ticker.C:
				s.l.Info("scraper ticker ticked")

				if err = s.ScrapeNewListings(ctx); err != nil {
					s.l.Error("failed to scrape new listings", logger.ErrAttr(err))
				}
			case <-ctx.Done():
				ticker.Stop()
				s.l.Info("scraper stopped", logger.ErrAttr(ctx.Err()))

				return
			}
		}
	}()

	s.l.Info("scraper service started")

	return nil
}

// ScrapeAllListings scrapes all listings by parameter.
func (s *Service) ScrapeAllListings(ctx context.Context) error {
	subscriptions, err := s.repo.GetAllSubscriptions(ctx)
	if err != nil {
		s.l.Error("failed to get subscriptions", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to get subscriptions")
	}

	if len(subscriptions) == 0 {
		s.l.Info("no subscriptions found")
		return nil
	}

	errCh := s.scrapeSubscriptions(ctx, subscriptions, s.scrapeAllListings)
	close(errCh)

	for err := range errCh {
		if err != nil {
			s.l.Error("scrape all listings error", logger.ErrAttr(err))
			return err
		}
	}

	s.l.Info("scraped all listings successfully")

	return nil
}

// ScrapeNewListings scrapes new listings for the past 24 hours.
func (s *Service) ScrapeNewListings(ctx context.Context) error {
	subscriptions, err := s.repo.GetAllSubscriptions(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get all subscriptions")
	}

	if len(subscriptions) == 0 {
		s.l.Info("no subscriptions found")
		return nil
	}

	errCh := s.scrapeSubscriptions(ctx, subscriptions, s.scrapeNewListings)
	close(errCh)

	for err := range errCh {
		if err != nil {
			return errors.Wrap(err, "scrape new listings error")
		}
	}

	return nil
}

// scrapeSubscriptions scrapes listings for each subscription using the provided scrape function.
func (s *Service) scrapeSubscriptions(
	ctx context.Context,
	subscriptions []ds.SubscriptionResponse,
	scrapeFunc func(context.Context, ds.SubscriptionResponse) error,
) chan error {
	var wg sync.WaitGroup

	errCh := make(chan error, len(subscriptions))

	for _, subscription := range subscriptions {
		wg.Add(1)

		go func(sub ds.SubscriptionResponse) {
			defer wg.Done()

			if err := scrapeFunc(ctx, sub); err != nil {
				errCh <- err
			}
		}(subscription)
	}

	wg.Wait()

	return errCh
}

// scrapeAllListings scrapes all listings for a subscription.
func (s *Service) scrapeAllListings(ctx context.Context, sub ds.SubscriptionResponse) error {
	params := subscriptionToParams(sub)

	listings, err := s.scrape(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to scrape listings by subscription ID "+sub.ID)
	}

	if len(listings) == 0 {
		s.l.Info("no listings found for subscription", logger.StringAttr("subscriptionID", sub.ID))
		return nil
	}

	for _, listing := range listings {
		if err = s.repo.UpsertListing(ctx, ds.UpsertListingRequest{
			ID:             listing.ID,
			SubscriptionID: sub.ID,
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
			return errors.Wrap(err, "failed to scrape listings for subscription ID "+sub.ID)
		}
	}

	return nil
}

// scrapeNewListings scrapes new listings for the past 24 hours for a subscription.
func (s *Service) scrapeNewListings(ctx context.Context, sub ds.SubscriptionResponse) error {
	params := subscriptionToParams(sub)
	params["sort"] = "renewDate_desc"
	params["date_limit"] = "1" // last 24h

	listings, err := s.scrape(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to scrape listings by subscription ID "+sub.ID)
	}

	if len(listings) == 0 {
		s.l.Info("no listings found for subscription", logger.StringAttr("subscriptionID", sub.ID))
		return nil
	}

	existListings, err := s.repo.GetListingsBySubscriptionID(ctx, sub.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get listings by subscription ID "+sub.ID)
	}

	isNeedSend := true
	if len(existListings) == 0 {
		// first time received a list of listings by subscription,
		// you don't need to send this to the user.
		isNeedSend = false
	}

	existingIDSet := make(map[string]struct{}, len(existListings))
	for _, existListing := range existListings {
		existingIDSet[existListing.ID] = struct{}{}
	}

	for _, listing := range listings {
		if _, exists := existingIDSet[listing.ID]; !exists {
			if err = s.repo.UpsertListing(ctx, ds.UpsertListingRequest{
				ID:             listing.ID,
				SubscriptionID: sub.ID,
				Title:          listing.Title,
				Price:          listing.Price,
				EngineVolume:   listing.EngineVolume,
				Transmission:   listing.Transmission,
				BodyType:       listing.BodyType,
				Mileage:        listing.Mileage,
				Location:       listing.Location,
				Link:           listing.Link,
				Date:           listing.Date,
				IsNeedSend:     isNeedSend, // it's important for send
			}); err != nil {
				return errors.Wrap(err, "failed to upsert listings for subscription ID "+sub.ID)
			}
		}
	}

	return nil
}

// scrape retrieves and processes car listings by parameters.
func (s *Service) scrape(ctx context.Context, params map[string]string) ([]polovniauto.Listing, error) {
	s.l.Info("scraping started")

	listings, err := s.paAdapter.GetNewListings(ctx, params)
	if err != nil {
		return []polovniauto.Listing{}, errors.Wrap(err, "failed to get new listings")
	}

	s.l.Info("scraping completed")

	return listings, nil
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
