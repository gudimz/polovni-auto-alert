package db

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	psql "github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db/sqlc_gen"
	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

var (
	ErrInvalidUUID = errors.New("invalid uuid")
)

// userToDB converts a ds.UserRequest to a psql.UpsertUserParams.
func userToDB(input ds.UserRequest) psql.UpsertUserParams {
	return psql.UpsertUserParams{
		ID:        input.ID,
		Username:  input.Username,
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}
}

// userFromDB converts a psql.User to a ds.UserResponse.
func userFromDB(input psql.User) ds.UserResponse {
	return ds.UserResponse{
		ID:        input.ID,
		Username:  input.Username,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		CreatedAt: input.CreatedAt.Time,
		UpdatedAt: input.UpdatedAt.Time,
	}
}

// subscriptionFromDB converts a psql.Subscription to a ds.SubscriptionResponse.
func subscriptionFromDB(input psql.Subscription) (ds.SubscriptionResponse, error) {
	id, err := pgUUIDToString(input.ID)
	if err != nil {
		return ds.SubscriptionResponse{}, err
	}

	return ds.SubscriptionResponse{
		ID:        id,
		UserID:    input.UserID,
		Brand:     input.Brand,
		Model:     input.Model,
		Chassis:   input.Chassis,
		PriceFrom: input.PriceFrom,
		PriceTo:   input.PriceTo,
		YearFrom:  input.YearFrom,
		YearTo:    input.YearTo,
		Region:    input.Region,
		CreatedAt: input.CreatedAt.Time,
		UpdatedAt: input.UpdatedAt.Time,
	}, nil
}

// subscriptionToDB converts a ds.SubscriptionRequest to a psql.CreateSubscriptionParams.
func subscriptionToDB(input ds.SubscriptionRequest) psql.CreateSubscriptionParams {
	return psql.CreateSubscriptionParams{
		UserID:    input.UserID,
		Brand:     input.Brand,
		Model:     input.Model,
		Chassis:   input.Chassis,
		PriceFrom: input.PriceFrom,
		PriceTo:   input.PriceTo,
		YearFrom:  input.YearFrom,
		YearTo:    input.YearTo,
		Region:    input.Region,
	}
}

// listingToDB converts a ds.UpsertListingRequest to psql.UpsertListingParams.
func listingToDB(input ds.UpsertListingRequest) (psql.UpsertListingParams, error) {
	subscriptionID, err := stringToPgUUID(input.SubscriptionID)
	if err != nil {
		return psql.UpsertListingParams{}, err
	}

	return psql.UpsertListingParams{
		ID:             input.ID,
		SubscriptionID: subscriptionID,
		Title:          input.Title,
		Price:          input.Price,
		EngineVolume:   input.EngineVolume,
		Transmission:   input.Transmission,
		BodyType:       input.BodyType,
		Mileage:        input.Mileage,
		Location:       input.Location,
		Link:           input.Link,
		Date:           timeToPgTimestamp(input.Date),
		IsNeedSend:     input.IsNeedSend,
	}, nil
}

// listingFromDB converts a ds.UpsertListingRequest to psql.UpsertListingParams.
func listingFromDB(input psql.Listing) (ds.ListingResponse, error) {
	subscriptionID, err := pgUUIDToString(input.SubscriptionID)
	if err != nil {
		return ds.ListingResponse{}, err
	}

	return ds.ListingResponse{
		ID:             input.ID,
		SubscriptionID: subscriptionID,
		Title:          input.Title,
		Price:          input.Price,
		EngineVolume:   input.EngineVolume,
		Transmission:   input.Transmission,
		BodyType:       input.BodyType,
		Mileage:        input.Mileage,
		Location:       input.Location,
		Link:           input.Link,
		Date:           input.Date.Time,
		IsNeedSend:     input.IsNeedSend,
		CreatedAt:      input.CreatedAt.Time,
		UpdatedAt:      input.UpdatedAt.Time,
	}, nil
}

// notificationToDB converts a ds.CreateNotificationRequest to psql.CreateNotificationParams.
func notificationToDB(input ds.CreateNotificationRequest) psql.CreateNotificationParams {
	return psql.CreateNotificationParams{
		ListingID: input.ListingID,
		Status:    statusToDB(input.Status),
		Reason:    input.Reason,
	}
}

// notificationFromDB converts a psql.Notification to ds.NotificationResponse.
func notificationFromDB(input psql.Notification) (ds.NotificationResponse, error) {
	id, err := pgUUIDToString(input.ID)
	if err != nil {
		return ds.NotificationResponse{}, err
	}

	return ds.NotificationResponse{
		ID:        id,
		ListingID: input.ListingID,
		Status:    statusFromDB(input.Status),
		Reason:    input.Reason,
		CreatedAt: input.CreatedAt.Time,
		UpdatedAt: input.UpdatedAt.Time,
	}, nil
}

// pgUUIDToString converts pgtype.UUID to string.
func pgUUIDToString(id pgtype.UUID) (string, error) {
	if !id.Valid {
		return "", ErrInvalidUUID
	}

	return (*uuid.UUID)(&id.Bytes).String(), nil
}

// stringToPgUUID converts string to pgtype.UUID.
func stringToPgUUID(id string) (pgtype.UUID, error) {
	var pgUUID pgtype.UUID

	err := pgUUID.Scan(id)
	if err != nil {
		return pgtype.UUID{}, ErrInvalidUUID
	}

	return pgUUID, nil
}

// timeToPgTimestamp converts time.Time to pgtype.Timestamp.
func timeToPgTimestamp(t time.Time) pgtype.Timestamp {
	if t.IsZero() {
		return pgtype.Timestamp{} //nolint:exhaustruct,nolintlint
	}

	return pgtype.Timestamp{Valid: true, Time: t} //nolint:exhaustruct,nolintlint
}

// statusToDB converts ds.NotificationStatus to psql.NullStatus.
func statusToDB(status ds.NotificationStatus) psql.Status {
	switch status {
	case ds.StatusSent:
		return psql.StatusSENT
	case ds.StatusFailed:
		return psql.StatusFAILED
	}

	return ""
}

// statusFromDB converts psql.Status to ds.NotificationStatus.
func statusFromDB(status psql.Status) ds.NotificationStatus {
	switch status {
	case psql.StatusSENT:
		return ds.StatusSent
	case psql.StatusFAILED:
		return ds.StatusFailed
	}

	return ""
}
