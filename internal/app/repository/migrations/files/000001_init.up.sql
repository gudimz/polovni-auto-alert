-- Create enum type for status
CREATE TYPE status AS ENUM ('SENT', 'FAILED');

-- Create users table
CREATE TABLE IF NOT EXISTS users
(
    id         BIGINT PRIMARY KEY,
    username   VARCHAR(256)            NOT NULL,
    first_name VARCHAR(256)            NOT NULL,
    last_name  VARCHAR(256)            NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL
);

-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions
(
    id         UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    BIGINT REFERENCES users (id) NOT NULL,
    brand      VARCHAR(256)                 NOT NULL,
    model      TEXT[]                       NOT NULL,
    chassis    TEXT[]                       NOT NULL,
    price_from VARCHAR(256)                 NOT NULL,
    price_to   VARCHAR(256)                 NOT NULL,
    year_from  VARCHAR(256)                 NOT NULL,
    year_to    VARCHAR(256)                 NOT NULL,
    region     TEXT[]                       NOT NULL,
    created_at TIMESTAMP DEFAULT now()      NOT NULL,
    updated_at TIMESTAMP DEFAULT now()      NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);

-- Create listings table
CREATE TABLE IF NOT EXISTS listings
(
    id              UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    listing_id      VARCHAR(256)            NOT NULL,
    subscription_id UUID REFERENCES subscriptions (id),
    title           VARCHAR(256)            NOT NULL,
    price           VARCHAR(256)            NOT NULL,
    engine_volume   VARCHAR(256)            NOT NULL,
    transmission    VARCHAR(256)            NOT NULL,
    body_type       VARCHAR(256)            NOT NULL,
    mileage         VARCHAR(256)            NOT NULL,
    location        VARCHAR(256)            NOT NULL,
    link            TEXT                    NOT NULL,
    date            TIMESTAMP               NOT NULL,
    is_need_send    BOOLEAN   DEFAULT FALSE NOT NULL,
    created_at      TIMESTAMP DEFAULT now() NOT NULL,
    updated_at      TIMESTAMP DEFAULT now() NOT NULL,

    CONSTRAINT listings_listing_id_subscription_id_unique UNIQUE (listing_id, subscription_id)

);

CREATE INDEX IF NOT EXISTS idx_listings_subscription_id ON listings (subscription_id);
CREATE INDEX IF NOT EXISTS idx_listings_is_need_send ON listings (is_need_send);


-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications
(
    id              UUID      DEFAULT gen_random_uuid() PRIMARY KEY,
    subscription_id UUID                    NOT NULL,
    listing_id      VARCHAR(256)            NOT NULL,
    status          status                  NOT NULL,
    reason          TEXT                    NOT NULL,
    created_at      TIMESTAMP DEFAULT now() NOT NULL,
    updated_at      TIMESTAMP DEFAULT now() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_listing_id ON notifications (listing_id);
