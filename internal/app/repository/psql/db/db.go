package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Register PostgreSQL driver
	pkgerrors "github.com/pkg/errors"

	"github.com/gudimz/polovni-auto-alert/internal/app/repository/migrations"
	psql "github.com/gudimz/polovni-auto-alert/internal/app/repository/psql/db/sqlc_gen"
	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type Repository struct {
	l       *logger.Logger
	cfg     *Config
	pool    *pgxpool.Pool
	queries *psql.Queries
}

func NewRepo(ctx context.Context, l *logger.Logger, cfg *Config) (*Repository, error) {
	hostPort := net.JoinHostPort(cfg.Host, cfg.Port)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s",
		cfg.User, cfg.Password, hostPort, cfg.DBName, cfg.SSLMode)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to pars postgres connection config")
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to connect to postgres")
	}

	return &Repository{
		l:       l,
		cfg:     cfg,
		pool:    pool,
		queries: psql.New(pool),
	}, nil
}

func (r *Repository) Connect() (*sql.DB, error) {
	hostPort := net.JoinHostPort(r.cfg.Host, r.cfg.Port)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s",
		r.cfg.User, r.cfg.Password, hostPort, r.cfg.DBName, r.cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to connect to database %s", dsn)
	}

	return db, nil
}

func (r *Repository) Migrate() error {
	db, err := r.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{}) //nolint:exhaustruct,nolintlint
	if err != nil {
		return pkgerrors.Wrap(err, "failed to create DB driver")
	}

	d, err := iofs.New(migrations.Files, "files")
	if err != nil {
		return pkgerrors.Wrap(err, "failed to create migration files")
	}

	m, err := migrate.NewWithInstance("iofs", d, r.cfg.DBName, driver)
	if err != nil {
		return pkgerrors.Wrapf(err, "failed to create migration instance")
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return pkgerrors.Wrap(err, "failed to run migrations")
	}

	r.l.Info("Database migrated successfully")

	return nil
}

func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
		r.l.Info("DB connection pool successfully closed")
	}
}

func (r *Repository) UpsertUser(ctx context.Context, request ds.UserRequest) (ds.UserResponse, error) {
	row, err := r.queries.UpsertUser(ctx, userToDB(request))
	if err != nil {
		return ds.UserResponse{}, pkgerrors.Wrap(err, "failed to upsert user to DB")
	}

	return userFromDB(row), nil
}

func (r *Repository) GetAllSubscriptions(ctx context.Context) ([]ds.SubscriptionResponse, error) {
	rows, err := r.queries.GetAllSubscriptions(ctx)
	if err != nil {
		return []ds.SubscriptionResponse{}, pkgerrors.Wrap(err, "failed to get all subscriptions from DB")
	}

	if len(rows) == 0 {
		return []ds.SubscriptionResponse{}, nil
	}

	subscriptions := make([]ds.SubscriptionResponse, 0, len(rows))

	for _, row := range rows {
		var subscription ds.SubscriptionResponse

		subscription, err = subscriptionFromDB(row)
		if err != nil {
			r.l.Warn("failed to convert subscription from DB", logger.ErrAttr(err))
			continue
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (r *Repository) CreateSubscription(
	ctx context.Context, sub ds.SubscriptionRequest,
) (ds.SubscriptionResponse, error) {
	row, err := r.queries.CreateSubscription(ctx, subscriptionToDB(sub))
	if err != nil {
		return ds.SubscriptionResponse{}, pkgerrors.Wrap(err, "failed to create subscription to DB")
	}

	return subscriptionFromDB(row)
}

func (r *Repository) UpsertListing(ctx context.Context, listing ds.UpsertListingRequest) error {
	req, err := listingToDB(listing)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to convert listing to DB")
	}

	if err = r.queries.UpsertListing(ctx, req); err != nil {
		return pkgerrors.Wrap(err, "failed to upsert listing to DB")
	}

	return nil
}

func (r *Repository) GetListingsBySubscriptionID(
	ctx context.Context, subscriptionID string,
) ([]ds.ListingResponse, error) {
	id, err := stringToPgUUID(subscriptionID)
	if err != nil {
		return []ds.ListingResponse{}, err
	}

	rows, err := r.queries.GetListingsBySubscriptionID(ctx, id)
	if err != nil {
		return []ds.ListingResponse{}, pkgerrors.Wrap(err, "failed to get listings by sub ID from DB")
	}

	if len(rows) == 0 {
		return []ds.ListingResponse{}, nil
	}

	listings := make([]ds.ListingResponse, 0, len(rows))

	for _, row := range rows {
		var listing ds.ListingResponse

		listing, err = listingFromDB(row)
		if err != nil {
			r.l.Warn("failed to convert listing from DB", logger.ErrAttr(err))
			continue
		}

		listings = append(listings, listing)
	}

	return listings, nil
}

func (r *Repository) GetListingsByIsNeedSend(ctx context.Context, isNeedSend bool) ([]ds.ListingResponse, error) {
	rows, err := r.queries.GetListingsByIsNeedSend(ctx, isNeedSend)
	if err != nil {
		return []ds.ListingResponse{}, pkgerrors.Wrap(err, "failed to get listings by is_need_send from DB")
	}

	listings := make([]ds.ListingResponse, 0, len(rows))

	for _, row := range rows {
		var listing ds.ListingResponse

		listing, err = listingFromDB(row)
		if err != nil {
			r.l.Warn("failed to convert listing from DB", logger.ErrAttr(err))
			continue
		}

		listings = append(listings, listing)
	}

	return listings, nil
}

func (r *Repository) CreateNotification(
	ctx context.Context, notification ds.CreateNotificationRequest,
) (ds.NotificationResponse, error) {
	row, err := r.queries.CreateNotification(ctx, notificationToDB(notification))
	if err != nil {
		return ds.NotificationResponse{}, pkgerrors.Wrap(err, "failed to create notification to DB")
	}

	return notificationFromDB(row)
}

func (r *Repository) GetSubscriptionsByUserID(ctx context.Context, userID int64) ([]ds.SubscriptionResponse, error) {
	rows, err := r.queries.GetSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return []ds.SubscriptionResponse{}, pkgerrors.Wrap(err, "failed to get subscriptions by user id from DB")
	}

	if len(rows) == 0 {
		return []ds.SubscriptionResponse{}, nil
	}

	subscriptions := make([]ds.SubscriptionResponse, 0, len(rows))

	for _, row := range rows {
		var subscription ds.SubscriptionResponse

		subscription, err = subscriptionFromDB(row)
		if err != nil {
			r.l.Warn("failed to convert subscription from DB", logger.ErrAttr(err))
			continue
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id string) (ds.SubscriptionResponse, error) {
	pgUUID, err := stringToPgUUID(id)
	if err != nil {
		return ds.SubscriptionResponse{}, err
	}

	row, err := r.queries.GetSubscriptionByID(ctx, pgUUID)
	if err != nil {
		return ds.SubscriptionResponse{}, pkgerrors.Wrap(err, "failed to get subscription by ID from DB")
	}

	return subscriptionFromDB(row)
}

func (r *Repository) DeleteListingsBySubscriptionIDs(ctx context.Context, ids []string) error {
	pgUUIDs := make([]pgtype.UUID, 0, len(ids))

	for _, id := range ids {
		pgUUID, err := stringToPgUUID(id)
		if err != nil {
			return err
		}

		pgUUIDs = append(pgUUIDs, pgUUID)
	}

	if err := r.queries.DeleteListingsBySubscriptionIDs(ctx, pgUUIDs); err != nil {
		return pkgerrors.Wrap(err, "failed to delete listings by sub ID to DB")
	}

	return nil
}

func (r *Repository) DeleteUserByID(ctx context.Context, id int64) error {
	if err := r.queries.DeleteUserByID(ctx, id); err != nil {
		return pkgerrors.Wrap(err, "failed to delete user by ID to DB")
	}

	return nil
}

func (r *Repository) DeleteSubscriptionsByUserID(ctx context.Context, userID int64) error {
	if err := r.queries.DeleteSubscriptionsByUserID(ctx, userID); err != nil {
		return pkgerrors.Wrap(err, "failed to delete subscriptions  by user ID to DB")
	}

	return nil
}

func (r *Repository) DeleteSubscriptionByID(ctx context.Context, id string) error {
	pgUUID, err := stringToPgUUID(id)
	if err != nil {
		return err
	}

	if err = r.queries.DeleteSubscriptionByID(ctx, pgUUID); err != nil {
		return pkgerrors.Wrap(err, "failed to delete subscription by ID from DB")
	}

	return nil
}
