package ds

import (
	"time"
)

type (
	UserRequest struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	UserResponse struct {
		ID        int64     `json:"id"`
		Username  string    `json:"username"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	SubscriptionRequest struct {
		UserID    int64    `json:"user_id"`
		Brand     string   `json:"brand"`
		Model     []string `json:"model"`
		Chassis   []string `json:"chassis"`
		PriceFrom string   `json:"price_from"`
		PriceTo   string   `json:"price_to"`
		YearFrom  string   `json:"year_from"`
		YearTo    string   `json:"year_to"`
		Region    []string `json:"region"`
	}

	SubscriptionResponse struct {
		ID        string    `json:"id"`
		UserID    int64     `json:"user_id"`
		Brand     string    `json:"brand"`
		Model     []string  `json:"model"`
		Chassis   []string  `json:"chassis"`
		PriceFrom string    `json:"price_from"`
		PriceTo   string    `json:"price_to"`
		YearFrom  string    `json:"year_from"`
		YearTo    string    `json:"year_to"`
		Region    []string  `json:"region"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	UpsertListingRequest struct {
		ListingID      string    `json:"listing_id"`
		SubscriptionID string    `json:"subscription_id"`
		Title          string    `json:"title"`
		Price          string    `json:"price"`
		EngineVolume   string    `json:"engine_volume"`
		Transmission   string    `json:"transmission"`
		BodyType       string    `json:"body_type"`
		Mileage        string    `json:"mileage"`
		Location       string    `json:"location"`
		Link           string    `json:"link"`
		Date           time.Time `json:"date"`
		IsNeedSend     bool      `json:"is_need_send"`
	}

	ListingResponse struct {
		ID             string    `json:"id"`
		ListingID      string    `json:"listing_id"`
		SubscriptionID string    `json:"subscription_id"`
		Title          string    `json:"title"`
		Price          string    `json:"price"`
		EngineVolume   string    `json:"engine_volume"`
		Transmission   string    `json:"transmission"`
		BodyType       string    `json:"body_type"`
		Mileage        string    `json:"mileage"`
		Location       string    `json:"location"`
		Link           string    `json:"link"`
		Date           time.Time `json:"date"`
		IsNeedSend     bool      `json:"is_need_send"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}
)
