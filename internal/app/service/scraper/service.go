package scraper

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/guregu/null"
	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	cache "github.com/gudimz/polovni-auto-alert/pkg/in_memory_storage"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

type Service struct {
	l           *logger.Logger
	repo        Repository
	paAdapter   PolovniAutoAdapter
	fetcher     Fetcher
	interval    time.Duration
	startOffset time.Duration
	workers     int
	// TODO: add job for updating the cache
	chassisList *cache.Storage[string, string]
}

// NewService creates a new Scraper Service instance.
func NewService(
	l *logger.Logger,
	repo Repository,
	paAdapter PolovniAutoAdapter,
	fetcher Fetcher,
	interval time.Duration,
	startOffset time.Duration,
	workers int,
) *Service {
	return &Service{
		l:           l,
		repo:        repo,
		paAdapter:   paAdapter,
		fetcher:     fetcher,
		interval:    interval,
		startOffset: startOffset,
		workers:     workers,
		chassisList: cache.New[string, string](),
	}
}

// Start begins the scraping process.
func (s *Service) Start(ctx context.Context) error {
	// set chassis list in cache
	chassis, err := s.fetcher.GetChassisFromJSON()
	if err != nil {
		return errors.Wrap(err, "failed to get chassis from json")
	}

	s.chassisList.SetBatch(chassis)

	s.l.Info(fmt.Sprintf("scraper service will start after: %v, interval: %v", s.startOffset, s.interval))

	// add a wait to not match the sending of notifications
	time.Sleep(s.startOffset)
	// to distinguish new listings from old ones,
	// we save all listings in the database when we start the app
	if err = s.ScrapeNewListings(ctx); err != nil {
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

	for err := range errCh {
		if err != nil {
			return errors.Wrap(err, "scrape new listings error")
		}
	}

	s.l.Info("scraped new listings successfully")

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
	tasks := make(chan ds.SubscriptionResponse, len(subscriptions))

	for i := 0; i < s.workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for sub := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
					if err := scrapeFunc(ctx, sub); err != nil {
						errCh <- err
					}
				}
			}
		}()
	}

	for _, sub := range subscriptions {
		tasks <- sub
	}

	close(tasks)

	wg.Wait()
	close(errCh)

	return errCh
}

// scrapeAllListings scrapes all listings for a subscription.
func (s *Service) scrapeAllListings(ctx context.Context, sub ds.SubscriptionResponse) error {
	params := s.subscriptionToParams(sub)

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
			ListingID:      listing.ID,
			SubscriptionID: sub.ID,
			Title:          listing.Title,
			Price:          listing.Price,
			NewPrice:       null.NewString("", false),
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
	params := s.subscriptionToParams(sub)
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

	existingListingIDSet := make(map[string]string, len(existListings))
	for _, existListing := range existListings {
		existingListingIDSet[existListing.ListingID] = existListing.Price
	}

	for _, listing := range listings {
		oldPrice, exists := existingListingIDSet[listing.ID]

		// check if the listing exists
		if exists && listing.Price == oldPrice {
			continue
		}

		req := ds.UpsertListingRequest{ //nolint:exhaustruct,nolintlint
			ListingID:      listing.ID,
			SubscriptionID: sub.ID,
			Title:          listing.Title,
			EngineVolume:   listing.EngineVolume,
			Transmission:   listing.Transmission,
			BodyType:       listing.BodyType,
			Mileage:        listing.Mileage,
			Location:       listing.Location,
			Link:           listing.Link,
			Date:           listing.Date,
			IsNeedSend:     isNeedSend, // it's important for send
		}

		if !exists {
			req.Price = listing.Price
		} else {
			req.Price = oldPrice
			req.NewPrice = null.NewString(listing.Price, listing.Price != "")
		}

		if err = s.repo.UpsertListing(ctx, req); err != nil {
			return errors.Wrap(err, "failed to upsert listings for subscription ID "+sub.ID)
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

// subscriptionToParams converts a subscription to a map of query parameters.
func (s *Service) subscriptionToParams(subscription ds.SubscriptionResponse) map[string]string {
	params := map[string]string{
		"brand":      subscription.Brand,
		"price_from": subscription.PriceFrom,
		"price_to":   subscription.PriceTo,
		"year_from":  subscription.YearFrom,
		"year_to":    subscription.YearTo,
		"showOldNew": "all",
	}

	if len(subscription.Model) > 0 {
		params["model[]"] = strings.Join(subscription.Model, ",")
	}

	if len(subscription.Region) > 0 {
		params["region[]"] = strings.Join(subscription.Region, ",")
	}

	if len(subscription.Chassis) > 0 {
		var mappedChassis []string

		// map chassis names to IDs
		for _, chassisName := range subscription.Chassis {
			if chassisID, exists := s.chassisList.Get(chassisName); exists {
				mappedChassis = append(mappedChassis, chassisID)
			}
		}

		params["chassis[]"] = strings.Join(mappedChassis, ",")
	}

	return params
}
