package notifier

import (
	"context"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

type (
	Repository interface {
		UpsertUser(ctx context.Context, request ds.UserRequest) (ds.UserResponse, error)
		CreateSubscription(ctx context.Context, sub ds.SubscriptionRequest) (ds.SubscriptionResponse, error)
		GetSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error)
		DeleteListingsBySubscriptionIDs(ctx context.Context, ids []string) error
		DeleteSubscriptionsByUserID(ctx context.Context, userID int64) error
		DeleteUserByID(ctx context.Context, id int64) error
		DeleteSubscriptionByID(ctx context.Context, id string) error
	}
)
