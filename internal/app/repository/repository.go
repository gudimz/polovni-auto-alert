package repository

import (
	"context"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

type DB interface {
	GetAllSubscriptions(ctx context.Context) ([]ds.SubscriptionResponse, error)
	CreateSubscription(ctx context.Context, sub ds.SubscriptionRequest) (ds.SubscriptionResponse, error)
	UpsertListing(ctx context.Context, listing ds.UpsertListingRequest) error
	GetListingsBySubscriptionID(ctx context.Context, subscriptionID string) ([]ds.ListingResponse, error)
	GetListingsByIsNeedSend(ctx context.Context, isNeedSend bool) ([]ds.ListingResponse, error)
	CreateNotification(ctx context.Context, notification ds.CreateNotificationRequest) (ds.NotificationResponse, error)
	GetSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error)
	DeleteListingsBySubscriptionIDs(ctx context.Context, ids []string) error
	DeleteSubscriptionsByUserID(ctx context.Context, userID int64) error
	DeleteUserByID(ctx context.Context, id int64)
	UpsertUser(ctx context.Context, request ds.UserRequest) (ds.UserResponse, error)
}
