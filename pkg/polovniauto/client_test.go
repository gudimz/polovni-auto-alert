package polovniauto

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type ClientTestSuite struct {
	suite.Suite
	client *Client
	server *httptest.Server
}

func (s *ClientTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`
			<article class="classified" data-classifiedid="1" data-price="2000€" data-renewdate="2023-10-10 10:10:10">
				<a class="ga-title" title="Best bmw" href="/auto-oglasi/1"></a>
				<div class="setInfo">
					<div class="top">2001. Limuzina</div>
					<div class="top">100000 km</div>
					<div class="top">Manual</div>
					<div class="bottom">2000 cm3</div>
				</div>
				<div class="city">Belgrade</div>
			</article>
		`))

		if err != nil {
			return
		}
	}))

	cfg := &Config{PageLimit: 2}
	lg := logger.NewLogger()
	baseURL, _ := url.Parse(s.server.URL)
	s.client = &Client{
		l:          lg,
		cfg:        cfg,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ClientTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *ClientTestSuite) TestClient_GetNewListings() {
	testCases := []struct {
		name      string
		params    map[string]string
		build     func() http.Handler
		want      []Listing
		expectErr error
	}{
		{
			name: "success",
			params: map[string]string{
				"brand":      "bmw",
				"model[]":    "m3,m5",
				"price_from": "1000",
				"price_to":   "3000",
			},
			build: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`
						<article class="classified" data-classifiedid="1" data-price="2000€" data-renewdate="2023-10-10 10:10:10">
							<a class="ga-title" title="Best bmw" href="/auto-oglasi/1"></a>
							<div class="setInfo">
								<div class="top">2001. Limuzina</div>
								<div class="top">100000 km</div>
								<div class="top">Manual</div>
								<div class="bottom">2000 cm3</div>
							</div>
							<div class="city">Belgrade</div>
						</article>
					`))
					if err != nil {
						return
					}
				})
			},
			want: []Listing{
				{
					ID:           "1",
					Title:        "Best bmw",
					Price:        "2000€",
					Year:         "2001",
					EngineVolume: "2000 cm3",
					Transmission: "Manual",
					BodyType:     "Limuzina",
					Mileage:      "100000 km",
					Location:     "Belgrade",
					Link:         s.server.URL + "/auto-oglasi/1",
					Date:         time.Date(2023, 10, 10, 10, 10, 10, 0, time.UTC),
				},
			},
		},
		{
			name: "unexpected status code",
			params: map[string]string{
				"brand":      "bmw",
				"model[]":    "m3,m5",
				"price_from": "1000",
				"price_to":   "3000",
			},
			build: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusForbidden)
				})
			},
			expectErr: ErrUnexpectedStatusCode,
		},
		{
			name: "empty listings",
			params: map[string]string{
				"brand":      "bmw",
				"model[]":    "m3,m5",
				"price_from": "1000",
				"price_to":   "3000",
			},
			build: func() http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(``))
					if err != nil {
						return
					}
				})
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s.server.Config.Handler = tc.build()

			got, err := s.client.GetNewListings(ctx, tc.params)

			switch {
			case tc.expectErr != nil:
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectErr, "expected error: %v, got: %v", tc.expectErr, err)
			default:
				s.Require().NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
