package worker

import (
	"context"
	"errors"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

var errCommon = errors.New("common error")

func (s *ServiceTestSuite) TestService_ProcessListings() {
	now := time.Now()
	listingID := uuid.NewString()
	subID := uuid.NewString()

	type testCase struct {
		mock      func(*testCase)
		name      string
		expectErr error
	}

	testCases := []testCase{
		{
			name: "success",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{
						ID:        subID,
						UserID:    1,
						Brand:     "bmw",
						Model:     []string{"m3", "m5"},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
				s.mockTgBot.EXPECT().SendMessage(gomock.Any()).
					Return(tgbotapi.Message{}, nil).
					Times(1)
				s.mockRepo.EXPECT().CreateNotification(gomock.Any(), ds.CreateNotificationRequest{
					SubscriptionID: subID,
					ListingID:      listingID,
					Status:         ds.StatusSent,
				}).Return(ds.NotificationResponse{
					ID:        uuid.NewString(),
					ListingID: listingID,
					Status:    ds.StatusSent,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil).
					Times(1)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingID,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2400€",
					Date:           now,
					IsNeedSend:     false,
				}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "get listings failed: common error",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{}, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
		{
			name: "get subscription failed: common error",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
		},
		{
			name: "send notification failed: common error",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{
						ID:        subID,
						UserID:    1,
						Brand:     "bmw",
						Model:     []string{"m3", "m5"},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
				s.mockTgBot.EXPECT().SendMessage(gomock.Any()).
					Return(tgbotapi.Message{}, errCommon).
					Times(1)
				s.mockRepo.EXPECT().CreateNotification(gomock.Any(), ds.CreateNotificationRequest{
					SubscriptionID: subID,
					ListingID:      listingID,
					Status:         ds.StatusFailed,
					Reason:         errCommon.Error(),
				}).Return(ds.NotificationResponse{
					ID:        uuid.NewString(),
					ListingID: listingID,
					Status:    ds.StatusFailed,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil).
					Times(1)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingID,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2400€",
					Date:           now,
					IsNeedSend:     true,
				}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "send notification failed: user blocked bot",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{
						ID:        subID,
						UserID:    1,
						Brand:     "bmw",
						Model:     []string{"m3", "m5"},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
				s.mockTgBot.EXPECT().SendMessage(gomock.Any()).
					Return(tgbotapi.Message{}, &tgbotapi.Error{
						Code: http.StatusForbidden,
					}).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), gomock.Any()).
					Return([]ds.SubscriptionResponse{
						{
							ID:        subID,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteSubscriptionsByUserID(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteUserByID(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "send notification failed: user blocked bot and failed remove all subscriptions",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{
						ID:        subID,
						UserID:    1,
						Brand:     "bmw",
						Model:     []string{"m3", "m5"},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
				s.mockTgBot.EXPECT().SendMessage(gomock.Any()).
					Return(tgbotapi.Message{}, &tgbotapi.Error{
						Code: http.StatusForbidden,
					}).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), gomock.Any()).
					Return([]ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
		},
		{
			name: "create notification failed: common error",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{
						ID:        subID,
						UserID:    1,
						Brand:     "bmw",
						Model:     []string{"m3", "m5"},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
				s.mockTgBot.EXPECT().SendMessage(gomock.Any()).
					Return(tgbotapi.Message{}, nil).
					Times(1)
				s.mockRepo.EXPECT().CreateNotification(gomock.Any(), ds.CreateNotificationRequest{
					SubscriptionID: subID,
					ListingID:      listingID,
					Status:         ds.StatusSent,
				}).Return(ds.NotificationResponse{}, errCommon).
					Times(1)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingID,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2400€",
					Date:           now,
					IsNeedSend:     false,
				}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "upsert listing failed: common error",
			mock: func(*testCase) {
				s.mockRepo.EXPECT().GetListingsByIsNeedSend(gomock.Any(), true).
					Return([]ds.ListingResponse{
						{
							ID:             listingID,
							SubscriptionID: subID,
							Title:          "Best bmw",
							Price:          "2400€",
							Date:           now,
							IsNeedSend:     true,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().GetSubscriptionByID(gomock.Any(), subID).
					Return(ds.SubscriptionResponse{
						ID:        subID,
						UserID:    1,
						Brand:     "bmw",
						Model:     []string{"m3", "m5"},
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
				s.mockTgBot.EXPECT().SendMessage(gomock.Any()).
					Return(tgbotapi.Message{}, nil).
					Times(1)
				s.mockRepo.EXPECT().CreateNotification(gomock.Any(), ds.CreateNotificationRequest{
					SubscriptionID: subID,
					ListingID:      listingID,
					Status:         ds.StatusSent,
				}).Return(ds.NotificationResponse{
					ID:        uuid.NewString(),
					ListingID: listingID,
					Status:    ds.StatusSent,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil).
					Times(1)
				s.mockRepo.EXPECT().UpsertListing(gomock.Any(), ds.UpsertListingRequest{
					ListingID:      listingID,
					SubscriptionID: subID,
					Title:          "Best bmw",
					Price:          "2400€",
					Date:           now,
					IsNeedSend:     false,
				}).
					Return(errCommon).
					Times(1)
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock(&tc)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err := s.svc.ProcessListings(ctx)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Assert().True(errors.Is(err, errCommon), "expected error: %v, got: %v", errCommon, err)
			default:
				s.Require().NoError(err)
			}

			cancel()
		})
	}
}

func (s *ServiceTestSuite) TestService_RemoveAllSubscriptionsByUserID() {
	now := time.Now()

	type testCase struct {
		mock      func(*testCase)
		name      string
		userID    int64
		expectErr error
	}

	testCases := []testCase{
		{
			name: "success",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "audi",
							Model:     []string{"a5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteUserByID(gomock.Any(), tc.userID).
					Return(nil).
					Times(1)
			},
			userID: 1,
		},
		{
			name: "get subscription from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
			userID:    1,
			expectErr: errCommon,
		},
		{
			name: "delete listings from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "audi",
							Model:     []string{"a5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), gomock.Any()).
					Return(errCommon).
					Times(1)
			},
			userID:    1,
			expectErr: errCommon,
		},
		{
			name: "delete subscriptions from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "audi",
							Model:     []string{"a5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return(errCommon).
					Times(1)
			},
			userID:    1,
			expectErr: errCommon,
		},
		{
			name: "delete user from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        uuid.NewString(),
							UserID:    1,
							Brand:     "audi",
							Model:     []string{"a5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteUserByID(gomock.Any(), tc.userID).
					Return(errCommon).
					Times(1)
			},
			userID:    1,
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock(&tc)

			err := s.svc.RemoveAllSubscriptionsByUserID(context.Background(), tc.userID)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Assert().True(errors.Is(err, errCommon), "expected error: %v, got: %v", errCommon, err)
			default:
				s.Require().NoError(err)
			}
		})
	}
}
