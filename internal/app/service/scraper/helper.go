package scraper

import (
	"strings"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

// subscriptionToParams converts a subscription to a map of query parameters.
func subscriptionToParams(subscription ds.SubscriptionResponse) map[string]string {
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

	return params
}
