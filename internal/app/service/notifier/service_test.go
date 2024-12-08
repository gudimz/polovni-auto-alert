package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/internal/pkg/ds"
	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

var errCommon = errors.New("common error")

type ServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *MockRepository
	mockPA   *MockPolovniAutoAdapter
	svc      *Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = NewMockRepository(s.ctrl)
	s.mockPA = NewMockPolovniAutoAdapter(s.ctrl)

	lg := logger.NewLogger()

	s.svc = NewService(
		lg,
		s.mockRepo,
		s.mockPA,
	)
	s.svc.carsList.SetBatch(map[string][]string{
		"bmw": {
			"m3",
			"m5",
		},
		"audi": {
			"a5",
		},
	})
	s.svc.carChassisList.SetBatch(map[string]string{
		"Limuzina": "277",
		"Pickup":   "2635",
	})
	s.svc.regionsList.SetBatch(map[string]string{"Beograd": "Beograd"})
}

func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

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
				s.Require().ErrorIsf(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Equal(tc.want, want)
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
				s.Require().ErrorIsf(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Equal(tc.want, want)
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
				s.Require().ErrorIsf(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.NoError(err)
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
				s.Require().ErrorIsf(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Equal(tc.want, want)
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
				s.Require().ErrorIsf(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ServiceTestSuite) TestService_UpdateCarList() {
	testCases := []struct {
		name      string
		mock      func()
		want      map[string][]string
		expectErr error
	}{
		{
			name: "empty list",
			mock: func() {
				s.mockPA.EXPECT().GetCarsList(gomock.Any()).
					Return(map[string][]string{}, nil).
					Times(1)
			},
			want: map[string][]string{ // no changes
				"bmw":  {"m3", "m5"},
				"audi": {"a5"},
			},
		},
		{
			name: "success",
			mock: func() {
				s.mockPA.EXPECT().GetCarsList(gomock.Any()).
					Return(map[string][]string{
						"bmw":  {"m3", "m5"},
						"audi": {"a5"},
					}, nil).
					Times(1)
			},
			want: map[string][]string{
				"bmw":  {"m3", "m5"},
				"audi": {"a5"},
			},
		},
		{
			name: "failed: common error",
			mock: func() {
				s.mockPA.EXPECT().GetCarsList(gomock.Any()).
					Return(nil, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := s.svc.UpdateCarList(ctx)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Require().Equal(tc.want, s.svc.carsList.CopyMap())
			}
		})
	}
}

func (s *ServiceTestSuite) TestService_CarRegionsList() {
	testCases := []struct {
		name      string
		mock      func()
		want      map[string]string
		expectErr error
	}{
		{
			name: "empty list",
			mock: func() {
				s.mockPA.EXPECT().GetRegionsList(gomock.Any()).
					Return(map[string]string{}, nil).
					Times(1)
			},
			want: map[string]string{ // no changes
				"Beograd": "Beograd",
			},
		},
		{
			name: "success",
			mock: func() {
				s.mockPA.EXPECT().GetRegionsList(gomock.Any()).
					Return(map[string]string{
						"Beograd":  "Beograd",
						"Novi Sad": "Novi Sad",
					}, nil).
					Times(1)
			},
			want: map[string]string{
				"Beograd":  "Beograd",
				"Novi Sad": "Novi Sad",
			},
		},
		{
			name: "failed: common error",
			mock: func() {
				s.mockPA.EXPECT().GetRegionsList(gomock.Any()).
					Return(nil, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := s.svc.UpdateCarRegionsList(ctx)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Require().Equal(tc.want, s.svc.regionsList.CopyMap())
			}
		})
	}
}

func (s *ServiceTestSuite) TestService_UpdateCarChassisList() {
	testCases := []struct {
		name      string
		mock      func()
		want      map[string]string
		expectErr error
	}{
		{
			name: "empty list",
			mock: func() {
				s.mockPA.EXPECT().GetCarChassisList(gomock.Any()).
					Return(map[string]string{}, nil).
					Times(1)
			},
			want: map[string]string{ // no changes
				"Limuzina": "277",
				"Pickup":   "2635",
			},
		},
		{
			name: "success",
			mock: func() {
				s.mockPA.EXPECT().GetCarChassisList(gomock.Any()).
					Return(map[string]string{
						"Karavan": "278",
						"Kupe":    "2633",
					}, nil).
					Times(1)
			},
			want: map[string]string{
				"Karavan": "278",
				"Kupe":    "2633",
			},
		},
		{
			name: "failed: common error",
			mock: func() {
				s.mockPA.EXPECT().GetCarChassisList(gomock.Any()).
					Return(nil, errCommon).
					Times(1)
			},
			expectErr: errCommon,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.mock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := s.svc.UpdateCarChassisList(ctx)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Require().Equal(tc.want, s.svc.carChassisList.CopyMap())
			}
		})
	}
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
