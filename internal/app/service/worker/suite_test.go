package worker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type ServiceTestSuite struct {
	suite.Suite
	ctrl      *gomock.Controller
	mockRepo  *MockRepository
	mockTgBot *MockTgBot
	svc       *Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	lg := logger.NewLogger()
	s.mockRepo = NewMockRepository(s.ctrl)
	s.mockTgBot = NewMockTgBot(s.ctrl)
	s.svc = NewService(lg, s.mockRepo, s.mockTgBot, 10*time.Second)
}

func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
