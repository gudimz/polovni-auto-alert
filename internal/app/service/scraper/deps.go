package scraper

import (
	"context"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

//go:generate mockgen -source=deps.go -destination=deps_mock.go -package=scraper
type (
	PolovniAutoAdapter interface {
		GetNewListings(ctx context.Context, params map[string]string) ([]polovniauto.Listing, error)
	}

	Repository interface {
		GetAllSubscriptions(ctx context.Context) ([]ds.SubscriptionResponse, error)
		UpsertListing(ctx context.Context, listing ds.UpsertListingRequest) error
		GetListingsBySubscriptionID(ctx context.Context, subscriptionID string) ([]ds.ListingResponse, error)
	}
)
