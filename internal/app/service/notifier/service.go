package notifier

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

// Service represents the notification service.
type Service struct {
	l    *logger.Logger
	repo Repository

	carsList    map[string][]string
	chassisList map[string]string
	regionsList map[string]string
}

//go:embed data/cars/cars.json
var carsJSON []byte

//go:embed data/chassis/chassis.json
var chassisJSON []byte

//go:embed data/regions/regions.json
var regionsJSON []byte

// NewService creates a new instance of the notification service.
func NewService(l *logger.Logger, repo Repository) (*Service, error) {
	svc := &Service{
		l:           l,
		repo:        repo,
		carsList:    make(map[string][]string),
		chassisList: make(map[string]string),
		regionsList: make(map[string]string),
	}

	if err := svc.loadData(); err != nil {
		return nil, errors.Wrap(err, "failed to load data from json")
	}

	return svc, nil
}

// UpsertUser creates or updates a user.
func (s *Service) UpsertUser(ctx context.Context, user ds.UserRequest) (ds.UserResponse, error) {
	u, err := s.repo.UpsertUser(ctx, user)
	if err != nil {
		s.l.Error("failed to create user", logger.ErrAttr(err))
		return ds.UserResponse{}, err
	}

	return u, nil
}

// CreateSubscription creates a new subscription.
func (s *Service) CreateSubscription(
	ctx context.Context, subscription ds.SubscriptionRequest,
) (ds.SubscriptionResponse, error) {
	sub, err := s.repo.CreateSubscription(ctx, subscription)
	if err != nil {
		s.l.Error("failed to create subscription", logger.ErrAttr(err))
		return ds.SubscriptionResponse{}, err
	}

	return sub, nil
}

// RemoveAllSubscriptionsByUserID removes all subscriptions and associated listings for a given user.
func (s *Service) RemoveAllSubscriptionsByUserID(ctx context.Context, userID int64) error {
	subscriptions, err := s.repo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		s.l.Error("failed to get subscriptions by user id", logger.ErrAttr(err))
		return err
	}

	ids := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		ids[i] = sub.ID
	}

	if err = s.repo.DeleteListingsBySubscriptionIDs(ctx, ids); err != nil {
		s.l.Error("failed to delete listings by subscription ids", logger.ErrAttr(err))
		return err
	}

	if err = s.repo.DeleteSubscriptionsByUserID(ctx, userID); err != nil {
		s.l.Error("failed to delete subscriptions by user id", logger.ErrAttr(err))
		return err
	}

	if err = s.repo.DeleteUserByID(ctx, userID); err != nil {
		s.l.Error("failed to delete user by id", logger.ErrAttr(err))
		return err
	}

	// TODO add transactions in future

	return nil
}

// GetAllSubscriptionsByUserID retrieves all subscriptions for a given user.
func (s *Service) GetAllSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error) {
	subscriptions, err := s.repo.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		s.l.Error("failed to get subscriptions by user id", logger.ErrAttr(err))
		return []ds.SubscriptionResponse{}, err
	}

	return subscriptions, err
}

// RemoveSubscriptionByID removes subscription and associated listings for a given subscription id.
func (s *Service) RemoveSubscriptionByID(ctx context.Context, id string) error {
	if err := s.repo.DeleteListingsBySubscriptionIDs(ctx, []string{id}); err != nil {
		s.l.Error("failed to delete listings by subscription ids", logger.ErrAttr(err))
		return err
	}

	if err := s.repo.DeleteSubscriptionByID(ctx, id); err != nil {
		s.l.Error("failed to delete subscription id", logger.ErrAttr(err))
		return err
	}

	// TODO add transactions in future

	return nil
}

// loadData loads the data from embedded JSON files into the service maps.
func (s *Service) loadData() error {
	if err := json.Unmarshal(carsJSON, &s.carsList); err != nil {
		return errors.Wrap(err, "failed to unmarshal cars json")
	}

	if err := json.Unmarshal(chassisJSON, &s.chassisList); err != nil {
		return errors.Wrap(err, "failed to load chassis data")
	}

	if err := json.Unmarshal(regionsJSON, &s.regionsList); err != nil {
		return errors.Wrap(err, "failed to load regions data")
	}

	return nil
}

func (s *Service) GetCarsList() map[string][]string {
	return s.carsList
}

func (s *Service) GetChassisList() map[string]string {
	return s.chassisList
}

func (s *Service) GetRegionsList() map[string]string {
	return s.regionsList
}
