package scraper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type ServiceTestSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	mockRepo         *MockRepository
	mockPpolovniAuto *MockPolovniAutoAdapter
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
		10*time.Second,
		5,
		map[string]string{
			"Limuzina": "277",
			"Pickup":   "2635",
		})
}

func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
