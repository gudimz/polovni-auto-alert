package notifier

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type ServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *MockRepository
	svc      *Service
}

func (s *ServiceTestSuite) SetupTest() {
	var err error

	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = NewMockRepository(s.ctrl)
	lg := logger.NewLogger()
	s.svc = NewService(
		lg,
		s.mockRepo,
		map[string][]string{
			"bmw": {
				"m3",
				"m5",
			},
			"audi": {
				"a5",
			},
		},
		map[string]string{"Beograd": "Beograd"},
		map[string]string{
			"Limuzina": "277",
			"Pickup":   "2635",
		},
	)
	s.Require().NoError(err)
}

func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
