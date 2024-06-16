package polovniauto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type Client struct {
	l          *logger.Logger
	cfg        *Config
	baseURL    *url.URL
	httpClient *http.Client
}

var ErrUnexpectedStatusCode = errors.New("unexpected status code")

func NewClient(l *logger.Logger, cfg *Config) *Client {
	baseURL, _ := url.Parse("https://www.polovniautomobili.com")

	return &Client{
		l:          l,
		cfg:        cfg,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second}, //nolint:nolintlint,mnd
	}
}

// Listing represents a car listing.
type Listing struct {
	ID           string
	Title        string
	Price        string
	Year         string
	EngineVolume string
	Transmission string
	BodyType     string
	Mileage      string
	Location     string
	Link         string
	Date         time.Time
}

// GetNewListings retrieves new car listings based on the provided parameters.
func (c *Client) GetNewListings(ctx context.Context, params map[string]string) ([]Listing, error) {
	var allListings []Listing

	page := 1

	for page != c.cfg.PageLimit {
		params["page"] = strconv.Itoa(page)

		uri := c.buildURL(params)
		c.l.Debug("visit: " + uri.String())

		bodyStr, err := c.fetchPage(ctx, uri)
		if err != nil {
			return nil, err
		}

		listings, err := c.parseListings(bodyStr)
		if err != nil {
			return nil, err
		}

		if len(listings) == 0 {
			break
		}

		allListings = append(allListings, listings...)
		page++
	}

	return allListings, nil
}

// buildURL constructs the URL with query parameters.
func (c *Client) buildURL(params map[string]string) *url.URL {
	rel := &url.URL{Path: "/auto-oglasi/pretraga"}
	u := c.baseURL.ResolveReference(rel)

	q := u.Query()

	for key, value := range params {
		switch key {
		case "model[]":
			for _, model := range strings.Split(value, ",") {
				q.Add("model[]", model)
			}
		case "region[]":
			for _, region := range strings.Split(value, ",") {
				q.Add("region[]", region)
			}
		default:
			q.Add(key, value)
		}
	}

	u.RawQuery = q.Encode()

	return u
}

// fetchPage retrieves the HTML content of the given URL.
func (c *Client) fetchPage(ctx context.Context, u *url.URL) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Decode HTML entities
	return html.UnescapeString(string(bodyBytes)), nil
}

// parseListings parses car listings from HTML.
func (c *Client) parseListings(bodyStr string) ([]Listing, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(bodyStr)))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var listings []Listing

	doc.Find("article.classified").Each(func(_ int, s *goquery.Selection) {
		id := strings.TrimSpace(s.AttrOr("data-classifiedid", "N/A"))
		title := strings.TrimSpace(s.Find("a.ga-title").AttrOr("title", "N/A"))
		price := strings.TrimSpace(s.AttrOr("data-price", "N/A"))
		dateStr := strings.TrimSpace(s.AttrOr("data-renewdate", "N/A"))
		engineVolume := strings.TrimSpace(s.Find("div.setInfo div.bottom").Eq(0).Text())
		transmission := strings.TrimSpace(s.Find("div.setInfo div.top").Eq(2).Text()) //nolint:nolintlint,mnd
		yearAndBodyType := strings.TrimSpace(s.Find("div.setInfo div.top").First().Text())
		mileage := strings.TrimSpace(s.Find("div.setInfo div.top").Eq(1).Text())
		location := strings.TrimSpace(s.Find("div.city").Text())
		link, _ := s.Find("a.ga-title").Attr("href")

		// Parse date string to time.Time
		var date time.Time

		date, err = time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			date = time.Time{} // Default to zero value if parsing fails
		}

		// Split year and body type
		year := "N/A"
		bodyType := "N/A"

		if parts := strings.SplitN(yearAndBodyType, ".", 2); len(parts) == 2 { //nolint:nolintlint,mnd
			year = strings.TrimSpace(parts[0])
			bodyType = strings.TrimSpace(parts[1])
		}

		// Ensure the link is a full URL
		if !strings.HasPrefix(link, "http") && c.baseURL != nil {
			link = c.baseURL.ResolveReference(&url.URL{Path: link}).String()
		}

		// Append parsed listing to the slice
		listings = append(listings, Listing{
			ID:           id,
			Title:        title,
			Price:        price,
			Year:         year,
			EngineVolume: engineVolume,
			Transmission: transmission,
			BodyType:     bodyType,
			Mileage:      mileage,
			Location:     location,
			Link:         link,
			Date:         date,
		})
	})

	return listings, nil
}
