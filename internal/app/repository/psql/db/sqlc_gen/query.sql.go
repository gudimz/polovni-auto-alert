// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const CreateNotification = `-- name: CreateNotification :one
INSERT INTO notifications (listing_id,
                           subscription_id,
                           status,
                           reason,
                           created_at,
                           updated_at)
VALUES ($1, $2, $3, $4, now(), now())
RETURNING id, subscription_id, listing_id, status, reason, created_at, updated_at
`

type CreateNotificationParams struct {
	ListingID      string      `json:"listing_id"`
	SubscriptionID pgtype.UUID `json:"subscription_id"`
	Status         Status      `json:"status"`
	Reason         string      `json:"reason"`
}

func (q *Queries) CreateNotification(ctx context.Context, arg CreateNotificationParams) (Notification, error) {
	row := q.db.QueryRow(ctx, CreateNotification,
		arg.ListingID,
		arg.SubscriptionID,
		arg.Status,
		arg.Reason,
	)
	var i Notification
	err := row.Scan(
		&i.ID,
		&i.SubscriptionID,
		&i.ListingID,
		&i.Status,
		&i.Reason,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const CreateSubscription = `-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id,
                           brand,
                           model,
                           chassis,
                           price_from,
                           price_to,
                           year_from,
                           year_to,
                           region,
                           created_at,
                           updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
RETURNING id, user_id, brand, model, chassis, price_from, price_to, year_from, year_to, region, created_at, updated_at
`

type CreateSubscriptionParams struct {
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

func (q *Queries) CreateSubscription(ctx context.Context, arg CreateSubscriptionParams) (Subscription, error) {
	row := q.db.QueryRow(ctx, CreateSubscription,
		arg.UserID,
		arg.Brand,
		arg.Model,
		arg.Chassis,
		arg.PriceFrom,
		arg.PriceTo,
		arg.YearFrom,
		arg.YearTo,
		arg.Region,
	)
	var i Subscription
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Brand,
		&i.Model,
		&i.Chassis,
		&i.PriceFrom,
		&i.PriceTo,
		&i.YearFrom,
		&i.YearTo,
		&i.Region,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const DeleteListingsBySubscriptionIDs = `-- name: DeleteListingsBySubscriptionIDs :exec
DELETE
FROM listings
WHERE subscription_id = ANY ($1::uuid[])
`

func (q *Queries) DeleteListingsBySubscriptionIDs(ctx context.Context, dollar_1 []pgtype.UUID) error {
	_, err := q.db.Exec(ctx, DeleteListingsBySubscriptionIDs, dollar_1)
	return err
}

const DeleteSubscriptionByID = `-- name: DeleteSubscriptionByID :exec
DELETE
FROM subscriptions
WHERE id = $1
`

func (q *Queries) DeleteSubscriptionByID(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, DeleteSubscriptionByID, id)
	return err
}

const DeleteSubscriptionsByUserID = `-- name: DeleteSubscriptionsByUserID :exec
DELETE
FROM subscriptions
WHERE user_id = $1
`

func (q *Queries) DeleteSubscriptionsByUserID(ctx context.Context, userID int64) error {
	_, err := q.db.Exec(ctx, DeleteSubscriptionsByUserID, userID)
	return err
}

const DeleteUserByID = `-- name: DeleteUserByID :exec
DELETE
FROM users
WHERE id = $1
`

func (q *Queries) DeleteUserByID(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, DeleteUserByID, id)
	return err
}

const GetAllSubscriptions = `-- name: GetAllSubscriptions :many
SELECT id,
       user_id,
       brand,
       model,
       chassis,
       price_from,
       price_to,
       year_from,
       year_to,
       region,
       created_at,
       updated_at
FROM subscriptions
`

func (q *Queries) GetAllSubscriptions(ctx context.Context) ([]Subscription, error) {
	rows, err := q.db.Query(ctx, GetAllSubscriptions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Subscription{}
	for rows.Next() {
		var i Subscription
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Brand,
			&i.Model,
			&i.Chassis,
			&i.PriceFrom,
			&i.PriceTo,
			&i.YearFrom,
			&i.YearTo,
			&i.Region,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetListingsByIsNeedSend = `-- name: GetListingsByIsNeedSend :many
SELECT id,
       listing_id,
       subscription_id,
       title,
       price,
       engine_volume,
       transmission,
       body_type,
       mileage,
       location,
       link,
       date,
       is_need_send,
       created_at,
       updated_at
FROM listings
WHERE is_need_send = $1
`

func (q *Queries) GetListingsByIsNeedSend(ctx context.Context, isNeedSend bool) ([]Listing, error) {
	rows, err := q.db.Query(ctx, GetListingsByIsNeedSend, isNeedSend)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Listing{}
	for rows.Next() {
		var i Listing
		if err := rows.Scan(
			&i.ID,
			&i.ListingID,
			&i.SubscriptionID,
			&i.Title,
			&i.Price,
			&i.EngineVolume,
			&i.Transmission,
			&i.BodyType,
			&i.Mileage,
			&i.Location,
			&i.Link,
			&i.Date,
			&i.IsNeedSend,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetListingsBySubscriptionID = `-- name: GetListingsBySubscriptionID :many
SELECT id,
       listing_id,
       subscription_id,
       title,
       price,
       engine_volume,
       transmission,
       body_type,
       mileage,
       location,
       link,
       date,
       is_need_send,
       created_at,
       updated_at
FROM listings
WHERE subscription_id = $1
`

func (q *Queries) GetListingsBySubscriptionID(ctx context.Context, subscriptionID pgtype.UUID) ([]Listing, error) {
	rows, err := q.db.Query(ctx, GetListingsBySubscriptionID, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Listing{}
	for rows.Next() {
		var i Listing
		if err := rows.Scan(
			&i.ID,
			&i.ListingID,
			&i.SubscriptionID,
			&i.Title,
			&i.Price,
			&i.EngineVolume,
			&i.Transmission,
			&i.BodyType,
			&i.Mileage,
			&i.Location,
			&i.Link,
			&i.Date,
			&i.IsNeedSend,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetSubscriptionByID = `-- name: GetSubscriptionByID :one
SELECT id,
       user_id,
       brand,
       model,
       chassis,
       price_from,
       price_to,
       year_from,
       year_to,
       region,
       created_at,
       updated_at
FROM subscriptions
WHERE id = $1
`

func (q *Queries) GetSubscriptionByID(ctx context.Context, id pgtype.UUID) (Subscription, error) {
	row := q.db.QueryRow(ctx, GetSubscriptionByID, id)
	var i Subscription
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Brand,
		&i.Model,
		&i.Chassis,
		&i.PriceFrom,
		&i.PriceTo,
		&i.YearFrom,
		&i.YearTo,
		&i.Region,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const GetSubscriptionsByUserID = `-- name: GetSubscriptionsByUserID :many
SELECT id,
       user_id,
       brand,
       model,
       chassis,
       price_from,
       price_to,
       year_from,
       year_to,
       region,
       created_at,
       updated_at
FROM subscriptions
WHERE user_id = $1
`

func (q *Queries) GetSubscriptionsByUserID(ctx context.Context, userID int64) ([]Subscription, error) {
	rows, err := q.db.Query(ctx, GetSubscriptionsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Subscription{}
	for rows.Next() {
		var i Subscription
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Brand,
			&i.Model,
			&i.Chassis,
			&i.PriceFrom,
			&i.PriceTo,
			&i.YearFrom,
			&i.YearTo,
			&i.Region,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const UpsertListing = `-- name: UpsertListing :exec
INSERT INTO listings (listing_id, subscription_id, title, price, engine_volume, transmission, body_type, mileage, location,
                      link, date, is_need_send, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now(), now())
ON CONFLICT (listing_id, subscription_id) DO UPDATE SET title         = EXCLUDED.title,
                                                price         = EXCLUDED.price,
                                                engine_volume = EXCLUDED.engine_volume,
                                                transmission  = EXCLUDED.transmission,
                                                body_type     = EXCLUDED.body_type,
                                                mileage       = EXCLUDED.mileage,
                                                location      = EXCLUDED.location,
                                                link          = EXCLUDED.link,
                                                date          = EXCLUDED.date,
                                                is_need_send  = EXCLUDED.is_need_send,
                                                updated_at    = now()
RETURNING id, listing_id, subscription_id, title, price, engine_volume, transmission, body_type, mileage, location, link, date, is_need_send, created_at, updated_at
`

type UpsertListingParams struct {
	ListingID      string           `json:"listing_id"`
	SubscriptionID pgtype.UUID      `json:"subscription_id"`
	Title          string           `json:"title"`
	Price          string           `json:"price"`
	EngineVolume   string           `json:"engine_volume"`
	Transmission   string           `json:"transmission"`
	BodyType       string           `json:"body_type"`
	Mileage        string           `json:"mileage"`
	Location       string           `json:"location"`
	Link           string           `json:"link"`
	Date           pgtype.Timestamp `json:"date"`
	IsNeedSend     bool             `json:"is_need_send"`
}

func (q *Queries) UpsertListing(ctx context.Context, arg UpsertListingParams) error {
	_, err := q.db.Exec(ctx, UpsertListing,
		arg.ListingID,
		arg.SubscriptionID,
		arg.Title,
		arg.Price,
		arg.EngineVolume,
		arg.Transmission,
		arg.BodyType,
		arg.Mileage,
		arg.Location,
		arg.Link,
		arg.Date,
		arg.IsNeedSend,
	)
	return err
}

const UpsertUser = `-- name: UpsertUser :one
INSERT INTO users(id,
                  username,
                  first_name,
                  last_name,
                  created_at,
                  updated_at)
VALUES ($1, $2, $3, $4, now(), now())
ON CONFLICT (id) DO UPDATE SET username   = EXCLUDED.username,
                               first_name = EXCLUDED.first_name,
                               last_name  = EXCLUDED.last_name,
                               updated_at = now()
RETURNING id, username, first_name, last_name, created_at, updated_at
`

type UpsertUserParams struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (q *Queries) UpsertUser(ctx context.Context, arg UpsertUserParams) (User, error) {
	row := q.db.QueryRow(ctx, UpsertUser,
		arg.ID,
		arg.Username,
		arg.FirstName,
		arg.LastName,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
