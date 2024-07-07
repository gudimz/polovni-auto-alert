package ds

import (
	"time"
)

type (
	CreateNotificationRequest struct {
		SubscriptionID string             `json:"subscription_id"`
		ListingID      string             `json:"listing_id"`
		Status         NotificationStatus `json:"status"`
		Reason         string             `json:"reason"`
	}

	NotificationResponse struct {
		ID        string             `json:"id"`
		ListingID string             `json:"listing_id"`
		Status    NotificationStatus `json:"status"`
		Reason    string             `json:"reason"`
		CreatedAt time.Time          `json:"created_at"`
		UpdatedAt time.Time          `json:"updated_at"`
	}

	NotificationStatus string
)

const (
	StatusSent   = NotificationStatus("SENT")
	StatusFailed = NotificationStatus("FAILED")
)
