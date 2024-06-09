-- name: UpsertUser :one
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
RETURNING *;

-- name: GetAllSubscriptions :many
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
FROM subscriptions;

-- name: CreateSubscription :one
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
RETURNING *;

-- name: UpsertListing :exec
INSERT INTO listings (id, subscription_id, title, price, engine_volume, transmission, body_type, mileage, location,
                      link, date, is_need_send, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now(), now())
ON CONFLICT (id) DO UPDATE SET subscription_id = EXCLUDED.subscription_id,
                               title           = EXCLUDED.title,
                               price           = EXCLUDED.price,
                               engine_volume   = EXCLUDED.engine_volume,
                               transmission    = EXCLUDED.transmission,
                               body_type       = EXCLUDED.body_type,
                               mileage         = EXCLUDED.mileage,
                               location        = EXCLUDED.location,
                               link            = EXCLUDED.link,
                               date            = EXCLUDED.date,
                               is_need_send    = EXCLUDED.is_need_send,
                               updated_at      = now()
RETURNING *;

-- name: GetListingsBySubscriptionID :many
SELECT id,
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
WHERE subscription_id = $1;

-- name: GetListingsByIsNeedSend :many
SELECT id,
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
WHERE is_need_send = $1;

-- name: CreateNotification :one
INSERT INTO notifications (listing_id,
                           status,
                           reason,
                           created_at,
                           updated_at)
VALUES ($1, $2, $3, now(), now())
RETURNING *;

-- name: GetSubscriptionsByUserID :many
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
WHERE user_id = $1;

-- name: GetSubscriptionByID :one
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
WHERE id = $1;

-- name: DeleteListingsBySubscriptionIDs :exec
DELETE
FROM listings
WHERE subscription_id = ANY ($1::uuid[]);

-- name: DeleteSubscriptionsByUserID :exec
DELETE
FROM subscriptions
WHERE user_id = $1;

-- name: DeleteSubscriptionByID :exec
DELETE
FROM subscriptions
WHERE id = $1;

-- name: DeleteUserByID :exec
DELETE
FROM users
WHERE id = $1;