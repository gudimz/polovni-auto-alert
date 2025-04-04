package scraper

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/polovniauto"
)

var errCommon = errors.New("common error")

type ServiceTestSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	mockRepo         *MockRepository
	mockPpolovniAuto *MockPolovniAutoAdapter
	mockFetcher      *MockFetcher
	svc              *Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	lg := logger.NewLogger()
	s.mockRepo = NewMockRepository(s.ctrl)
	s.mockPpolovniAuto = NewMockPolovniAutoAdapter(s.ctrl)
	s.svc = NewService(
		lg,
		s.mockRepo,
		s.mockPpolovniAuto,
		s.mockFetcher,
		10*time.Second,
		5,
	)

	s.svc.chassisList.SetBatch(map[string]string{
		"Limuzina": "277",
		"Pickup":   "2635",
	})
}

func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ServiceTestSuite) TestService_ScrapeAllListings() {
	now := time.Now()
	listingID := uuid.NewString()
	subID := uuid.NewString()

	testCases := []struct {
		name      string
		mock      func()
		expectErr error
	}{
		{
			name: "success",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{
						{
							ID:    listingID,
							Title: "Best bmw",
							Price: "2000€",
							Year:  "2001",
							Date:  now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingID,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2000€",
					Date:           now,
					IsNeedSend:     false,
				}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "success no subscriptions find",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{}, nil).
					Times(1)
			},
		},
		{
			name: "success no listings find",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{}, nil).
					Times(1)
			},
		},
		{
			name: "get all subscriptions failed: common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
		{
			name: "get listings failed: common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{}, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
		{
			name: "failed to upsert listing to DB: common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{
						{
							ID:    listingID,
							Title: "Best bmw",
							Price: "2000€",
							Year:  "2001",
							Date:  now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingID,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2000€",
					Date:           now,
					IsNeedSend:     false,
				}).
					Return(errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err := s.svc.ScrapeAllListings(ctx)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
			}

			cancel()
		})
	}
}

func (s *ServiceTestSuite) TestService_ScrapeNewListings() {
	now := time.Now()
	listingIDExist := "1"
	listingIDNotExist := "2"
	subID := uuid.NewString()

	testCases := []struct {
		name      string
		mock      func()
		expectErr error
	}{
		{
			name: "success",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"sort":       "renewDate_desc",
					"date_limit": "1",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{
						{
							ID:    listingIDExist,
							Title: "Best audi",
							Price: "2400€",
							Year:  "2002",
							Date:  now,
						},
						{
							ID:    listingIDNotExist,
							Title: "Best bmw",
							Price: "2000€",
							Year:  "2001",
							Date:  now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetListingsBySubscriptionID(gomock.Any(), subID).
					Return([]ds.ListingResponse{
						{
							ID:             uuid.NewString(),
							ListingID:      listingIDExist,
							SubscriptionID: subID,
							Title:          "Best audi",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     false,
						},
					}, nil)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingIDNotExist,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2000€",
					Date:           now,
					IsNeedSend:     true,
				}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "success no subscriptions find",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{}, nil).
					Times(1)
			},
		},
		{
			name: "success no listings find",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"sort":       "renewDate_desc",
					"date_limit": "1",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{}, nil).
					Times(1)
			},
		},
		{
			name: "success: first time, no need send listings",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"sort":       "renewDate_desc",
					"date_limit": "1",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{
						{
							ID:    listingIDNotExist,
							Title: "Best bmw in the world",
							Price: "2300€",
							Year:  "2004",
							Date:  now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetListingsBySubscriptionID(gomock.Any(), subID).
					Return([]ds.ListingResponse{}, nil)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingIDNotExist,
					SubscriptionID: subID,
					Title:          "Best bmw in the world",
					Price:          "2300€",
					Date:           now,
					IsNeedSend:     false,
				}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "get all subscriptions failed: common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
		{
			name: "get listings failed: common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"sort":       "renewDate_desc",
					"date_limit": "1",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{}, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
		{
			name: "upsert listings failed: get listings common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"sort":       "renewDate_desc",
					"date_limit": "1",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{
						{
							ID:    listingIDExist,
							Title: "Best audi",
							Price: "2400€",
							Year:  "2002",
							Date:  now,
						},
						{
							ID:    listingIDNotExist,
							Title: "Best bmw",
							Price: "2000€",
							Year:  "2001",
							Date:  now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetListingsBySubscriptionID(gomock.Any(), subID).
					Return([]ds.ListingResponse{}, errCommon)
			},
			expectErr: errCommon,
		},
		{
			name: "upsert listings failed: common error",
			mock: func() {
				s.mockRepo.EXPECT().GetAllSubscriptions(gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							PriceFrom: "1000",
							PriceTo:   "3000",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockPpolovniAuto.EXPECT().GetNewListings(gomock.Any(), map[string]string{
					"brand":      "bmw",
					"model[]":    "m3,m5",
					"price_from": "1000",
					"price_to":   "3000",
					"year_from":  "",
					"year_to":    "",
					"sort":       "renewDate_desc",
					"date_limit": "1",
					"showOldNew": "all",
				}).
					Return([]polovniauto.Listing{
						{
							ID:    listingIDExist,
							Title: "Best audi",
							Price: "2400€",
							Year:  "2002",
							Date:  now,
						},
						{
							ID:    listingIDNotExist,
							Title: "Best bmw",
							Price: "2000€",
							Year:  "2001",
							Date:  now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetListingsBySubscriptionID(gomock.Any(), subID).
					Return([]ds.ListingResponse{
						{
							ID:             uuid.NewString(),
							ListingID:      listingIDExist,
							SubscriptionID: subID,
							Title:          "Best audi",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     false,
						},
					}, nil)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingIDNotExist,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2000€",
					Date:           now,
					IsNeedSend:     true,
				}).
					Return(errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err := s.svc.ScrapeNewListings(ctx)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
			}

			cancel()
		})
	}
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
