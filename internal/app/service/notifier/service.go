package notifier

import (
	"context"

	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	cache "github.com/gudimz/polovni-auto-alert/pkg/in_memory_storage"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

// Service represents the notification service.
type Service struct {
	l       *logger.Logger
	repo    Repository
	fetcher Fetcher
	// TODO: add job for updating the cache
	carsList    *cache.Storage[string, []string]
	chassisList *cache.Storage[string, string]
	regionsList *cache.Storage[string, string]
}

// NewService creates a new instance of the notification service.
func NewService(l *logger.Logger, repo Repository, fetcher Fetcher) *Service {
	return &Service{
		l:           l,
		repo:        repo,
		fetcher:     fetcher,
		carsList:    cache.New[string, []string](),
		chassisList: cache.New[string, string](),
		regionsList: cache.New[string, string](),
	}
}

// Start begins the notification process.
func (s *Service) Start() error {
	// set cars list in cache
	cars, err := s.fetcher.GetCarsFromJSON()
	if err != nil {
		return errors.Wrap(err, "failed to get cars from json")
	}

	s.carsList.SetBatch(cars)

	// set chassis list in cache
	chassis, err := s.fetcher.GetChassisFromJSON()
	if err != nil {
		return errors.Wrap(err, "failed to get chassis from json")
	}

	s.chassisList.SetBatch(chassis)

	// set regions list in cache
	regions, err := s.fetcher.GetRegionsFromJSON()
	if err != nil {
		return errors.Wrap(err, "failed to get regions from json")
	}

	s.regionsList.SetBatch(regions)

	s.l.Info("notifier service started")

	return nil
}

// UpsertUser creates or updates a user.
func (s *Service) UpsertUser(ctx context.Context, user ds.UserRequest) (ds.UserResponse, error) {
	lg := logger.L(ctx).With(
		logger.Int64Attr("user_id", user.ID),
	)

	u, err := s.repo.UpsertUser(ctx, user)
	if err != nil {
		lg.Error("failed to create user", logger.ErrAttr(err))
		return ds.UserResponse{}, errors.Wrap(err, "failed to create user")
	}

	return u, nil
}

// CreateSubscription creates a new subscription.
func (s *Service) CreateSubscription(
	ctx context.Context, subscription ds.SubscriptionRequest,
) (ds.SubscriptionResponse, error) {
	lg := s.l.With(logger.Int64Attr("user_id", subscription.UserID),
		logger.StringAttr("brand", subscription.Brand),
		logger.AnyAttr("models", subscription.Model),
	)

	sub, err := s.repo.CreateSubscription(ctx, subscription)
	if err != nil {
		lg.Error("failed to create subscription", logger.ErrAttr(err))
		return ds.SubscriptionResponse{}, errors.Wrap(err, "failed to create subscription")
	}

	return sub, nil
}

// RemoveAllSubscriptionsByUserID removes all subscriptions and associated listings for a given user.
func (s *Service) RemoveAllSubscriptionsByUserID(ctx context.Context, userID int64) error {
	lg := s.l.With(logger.Int64Attr("user_id", userID))

	subscriptions, err := s.repo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		lg.Error("failed to get subscriptions by user id", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to get subscriptions by user id")
	}

	ids := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		ids[i] = sub.ID
	}

	if err = s.repo.DeleteListingsBySubscriptionIDs(ctx, ids); err != nil {
		lg.Error("failed to delete listings by subscription ids", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to delete listings by subscription ids")
	}

	if err = s.repo.DeleteSubscriptionsByUserID(ctx, userID); err != nil {
		lg.Error("failed to delete subscriptions by user id", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to delete subscriptions by user")
	}

	if err = s.repo.DeleteUserByID(ctx, userID); err != nil {
		lg.Error("failed to delete user by id", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to delete user by id")
	}

	// TODO add transactions in future

	return nil
}

// GetAllSubscriptionsByUserID retrieves all subscriptions for a given user.
func (s *Service) GetAllSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error) {
	lg := s.l.With(logger.Int64Attr("user_id", userID))

	subscriptions, err := s.repo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		lg.Error("failed to get subscriptions by user id", logger.ErrAttr(err))
		return []ds.SubscriptionResponse{}, errors.Wrap(err, "failed to get subscriptions by user id")
	}

	return subscriptions, nil
}

// RemoveSubscriptionByID removes subscription and associated listings for a given subscription id.
func (s *Service) RemoveSubscriptionByID(ctx context.Context, id string) error {
	lg := s.l.With(logger.StringAttr("subscription_id", id))
	if err := s.repo.DeleteListingsBySubscriptionIDs(ctx, []string{id}); err != nil {
		lg.Error("failed to delete listings by subscription ids", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to delete listings by subscription ids")
	}

	if err := s.repo.DeleteSubscriptionByID(ctx, id); err != nil {
		lg.Error("failed to delete subscription by id", logger.ErrAttr(err))
		return errors.Wrap(err, "failed to delete subscription by id")
	}

	// TODO add transactions in future

	return nil
}

// GetCarBrandsList retrieves the list of car brands.
func (s *Service) GetCarBrandsList() []string {
	return s.carsList.Keys()
}

// GetCarModelsList retrieves the list of car models for a given brand.
func (s *Service) GetCarModelsList(brand string) ([]string, bool) {
	return s.carsList.Get(brand)
}

// GetCarChassisList retrieves the list of car body types.
func (s *Service) GetCarChassisList() map[string]string {
	return s.chassisList.CopyMap()
}

// GetRegionsList retrieves the list of regions.
func (s *Service) GetRegionsList() map[string]string {
	return s.regionsList.CopyMap()
}
