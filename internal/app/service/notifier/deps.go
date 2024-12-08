package notifier

import (
	"context"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

//go:generate mockgen -source=deps.go -destination=deps_mock.go -package=notifier
type (
	Repository interface {
		UpsertUser(context.Context, ds.UserRequest) (ds.UserResponse, error)
		CreateSubscription(context.Context, ds.SubscriptionRequest) (ds.SubscriptionResponse, error)
		GetSubscriptionsByUserID(context.Context /*userID*/, int64) ([]ds.SubscriptionResponse, error)
		DeleteListingsBySubscriptionIDs(context.Context /*ids*/, []string) error
		DeleteSubscriptionsByUserID(context.Context /*userID*/, int64) error
		DeleteUserByID(context.Context /*id*/, int64) error
		DeleteSubscriptionByID(context.Context /*id*/, string) error
	}

	PolovniAutoAdapter interface {
		GetCarsList(context.Context) (map[string][]string, error)
		GetCarChassisList(context.Context) (map[string]string, error)
		GetRegionsList(context.Context) (map[string]string, error)
	}
)
