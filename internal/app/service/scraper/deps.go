package scraper

import (
	"context"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

//go:generate mockgen -source=deps.go -destination=deps_mock.go -package=scraper
type (
	PolovniAutoAdapter interface {
		GetNewListings(context.Context, map[string]string) ([]polovniauto.Listing, error)
		GetCarChassisList(context.Context) (map[string]string, error)
	}

	Repository interface {
		GetAllSubscriptions(context.Context) ([]ds.SubscriptionResponse, error)
		UpsertListing(context.Context, ds.UpsertListingRequest) error
		GetListingsBySubscriptionID(context.Context, string) ([]ds.ListingResponse, error)
	}
)
