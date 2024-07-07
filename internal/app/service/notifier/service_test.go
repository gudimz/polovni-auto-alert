package notifier

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
)

var errCommon = errors.New("common error")

func (s *ServiceTestSuite) TestService_UpsertUser() {
	now := time.Now()

	type testCase struct {
		name      string
		mock      func(*testCase)
		user      ds.UserRequest
		want      ds.UserResponse
		expectErr error
	}

	testCases := []testCase{
		{
			name: "success",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().UpsertUser(gomock.Any(), tc.user).
					Return(tc.want, nil).
					Times(1)
			},
			user: ds.UserRequest{
				ID:        1,
				Username:  "testuser",
				FirstName: "test",
				LastName:  "user",
			},
			want: ds.UserResponse{
				ID:        1,
				Username:  "testuser",
				FirstName: "test",
				LastName:  "user",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "upsert user to DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().UpsertUser(gomock.Any(), tc.user).
					Return(ds.UserResponse{}, errCommon).
					Times(1)
			},
			user: ds.UserRequest{
				ID:        1,
				Username:  "testuser",
				FirstName: "test",
				LastName:  "user",
			},
			expectErr: errCommon,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock(&tc)

			want, err := s.svc.UpsertUser(context.Background(), tc.user)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Assert().True(errors.Is(err, errCommon), "expected error: %v, got: %v", errCommon, err)
			default:
				s.Require().NoError(err)
				s.Assert().Equal(tc.want, want)
			}
		})
	}
}

func (s *ServiceTestSuite) TestService_CreateSubscription() {
	now := time.Now()

	type testCase struct {
		mock         func(*testCase)
		name         string
		subscription ds.SubscriptionRequest
		want         ds.SubscriptionResponse
		expectErr    error
	}

	testCases := []testCase{
		{
			name: "success",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().CreateSubscription(gomock.Any(), tc.subscription).
					Return(tc.want, nil).
					Times(1)
			},
			subscription: ds.SubscriptionRequest{
				UserID: 1,
				Brand:  "bmw",
				Model:  []string{"m3", "m5"},
			},
			want: ds.SubscriptionResponse{
				ID:        uuid.NewString(),
				UserID:    1,
				Brand:     "bmw",
				Model:     []string{"m3", "m5"},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "create subscription to DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().CreateSubscription(gomock.Any(), tc.subscription).
					Return(ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
			subscription: ds.SubscriptionRequest{
				UserID: 1,
				Brand:  "bmw",
				Model:  []string{"m3", "m5"},
			},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock(&tc)

			want, err := s.svc.CreateSubscription(context.Background(), tc.subscription)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Assert().True(errors.Is(err, errCommon), "expected error: %v, got: %v", errCommon, err)
			default:
				s.Require().NoError(err)
				s.Assert().Equal(tc.want, want)
			}
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

func (s *ServiceTestSuite) TestService_GetAllSubscriptionsByUserID() {
	now := time.Now()
	id := uuid.NewString()

	type testCase struct {
		mock      func(*testCase)
		name      string
		userID    int64
		want      []ds.SubscriptionResponse
		expectErr error
	}

	testCases := []testCase{
		{
			name: "success",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{
						{
							ID:        id,
							UserID:    1,
							Brand:     "bmw",
							Model:     []string{"m3", "m5"},
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).
					Times(1)
			},
			userID: 1,
			want: []ds.SubscriptionResponse{
				{
					ID:        id,
					UserID:    1,
					Brand:     "bmw",
					Model:     []string{"m3", "m5"},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
		{
			name: "get subscription from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().GetSubscriptionsByUserID(gomock.Any(), tc.userID).
					Return([]ds.SubscriptionResponse{}, errCommon).
					Times(1)
			},
			userID:    1,
			want:      []ds.SubscriptionResponse{},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock(&tc)

			want, err := s.svc.GetAllSubscriptionsByUserID(context.Background(), tc.userID)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Assert().True(errors.Is(err, errCommon), "expected error: %v, got: %v", errCommon, err)
			default:
				s.Require().NoError(err)
				s.Assert().Equal(tc.want, want)
			}
		})
	}
}

func (s *ServiceTestSuite) TestService_RemoveSubscriptionByID() {
	id := uuid.NewString()

	type testCase struct {
		mock      func(*testCase)
		name      string
		id        string
		expectErr error
	}

	testCases := []testCase{
		{
			name: "success",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), []string{tc.id}).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteSubscriptionByID(gomock.Any(), tc.id).
					Return(nil).
					Times(1)
			},
			id: id,
		},
		{
			name: "delete listings from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), []string{tc.id}).
					Return(errCommon).
					Times(1)
			},
			id:        id,
			expectErr: errCommon,
		},
		{
			name: "delete subscription from DB failed: common error",
			mock: func(tc *testCase) {
				s.mockRepo.EXPECT().DeleteListingsBySubscriptionIDs(gomock.Any(), []string{tc.id}).
					Return(nil).
					Times(1)
				s.mockRepo.EXPECT().DeleteSubscriptionByID(gomock.Any(), tc.id).
					Return(errCommon).
					Times(1)
			},
			id:        id,
			expectErr: errCommon,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock(&tc)

			err := s.svc.RemoveSubscriptionByID(context.Background(), tc.id)

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

func (s *ServiceTestSuite) TestService_GetCarsList() {
	want := map[string][]string{
		"bmw":  {"m3", "m6"},
		"audi": {"a1", "a3"},
	}

	s.svc.carsList = want

	got := s.svc.GetCarsList()
	s.Equal(want, got)
}

func (s *ServiceTestSuite) TestService_GetChassisList() {
	want := map[string]string{
		"Pickup": "2635",
		"Kupe":   "2633",
	}

	s.svc.chassisList = want

	got := s.svc.GetChassisList()
	s.Assert().Equal(want, got)
}

func (s *ServiceTestSuite) TestService_GetRegionsList() {
	want := map[string]string{
		"Beograd":   "Beograd",
		"Vojvodina": "Vojvodina",
	}

	s.svc.regionsList = want

	got := s.svc.GetRegionsList()
	s.Assert().Equal(want, got)
}
