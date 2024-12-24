package polovniauto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
	"github.com/gudimz/polovni-auto-alert/pkg/utils"
)

type Client struct {
	l          *logger.Logger
	cfg        *Config
	baseURL    *url.URL
	httpClient *http.Client
}

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

const (
	urlPA = "https://www.polovniautomobili.com"

	httpTimeout           = 60 * time.Second
	maxIdleConnects       = 100
	idleConnTimeout       = 90 * time.Second
	maxConnPerHost        = 30
	tlsHandshakeTimeout   = 15 * time.Second
	expectContinueTimeout = 1 * time.Second
	responseHeaderTimeout = 15 * time.Second
	dialerTimeout         = 30 * time.Second
	dialerKeepAlive       = 30 * time.Second

	delay          = 1 * time.Second
	maxRandomDelay = 3
)

func NewClient(l *logger.Logger, cfg *Config) *Client {
	baseURL, _ := url.Parse(urlPA)

	return &Client{
		l:       l,
		cfg:     cfg,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				MaxIdleConns:          maxIdleConnects,
				MaxConnsPerHost:       maxConnPerHost,
				IdleConnTimeout:       idleConnTimeout,
				TLSHandshakeTimeout:   tlsHandshakeTimeout,
				ExpectContinueTimeout: expectContinueTimeout,
				ResponseHeaderTimeout: responseHeaderTimeout,
				DialContext: (&net.Dialer{ //nolint:exhaustruct,nolintlint
					Timeout:   dialerTimeout,
					KeepAlive: dialerKeepAlive,
				}).DialContext,
			},
		},
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
			return []Listing{}, err
		}

		listings, err := c.parseListings(bodyStr)
		if err != nil {
			return []Listing{}, err
		}

		if len(listings) == 0 {
			break
		}

		allListings = append(allListings, listings...)
		page++

		// random delay to avoid being blocked
		randomDelay := time.Duration(rand.IntN(maxRandomDelay)+1) * time.Second //nolint:gosec,nolintlint
		time.Sleep(randomDelay)
	}

	return allListings, nil
}

// GetCarsList retrieves the list of car brands and models.
func (c *Client) GetCarsList(ctx context.Context) (map[string][]string, error) {
	ctx, cancel := chromedp.NewRemoteAllocator(ctx, c.cfg.ChromeWSURL)
	defer cancel()

	ctxTask, taskCancel := chromedp.NewContext(ctx)
	defer taskCancel()

	c.l.Info("loading the site for cars list")

	if err := chromedp.Run(
		ctxTask,
		chromedp.Navigate(urlPA),
		chromedp.WaitVisible(`#brand`, chromedp.ByID), // wait for the brand list to load
		chromedp.Sleep(delay),
	); err != nil {
		return nil, fmt.Errorf("error loading the site for cars list: %w", err)
	}

	var brands []map[string]string
	if err := chromedp.Run(
		ctxTask,
		chromedp.Evaluate(`Array.from(document.querySelectorAll("#brand option"))
        .map(option => ({ name: option.textContent.trim(), id: option.value }))`, &brands),
	); err != nil {
		return nil, fmt.Errorf("error getting brands for cars list: %w", err)
	}

	c.l.Debug(fmt.Sprintf("found the brands: %d", len(brands)))

	modelsAndBrands := make(map[string][]string)

	// iterate over brands and collect models for each
	for _, brand := range brands {
		if brand["id"] == "" {
			continue
		}

		c.l.Debug("processing brand: " + brand["name"])

		models := []string{}
		if err := chromedp.Run(
			ctxTask,
			// clear the model list
			chromedp.Evaluate(`document.querySelector("#model").innerHTML = ""`, nil),
			// select the brand
			chromedp.SetValue(`#brand`, brand["id"], chromedp.ByID),
			// generate change event
			chromedp.Evaluate(`document.querySelector("#brand").dispatchEvent(new Event('change'))`, nil),
			// wait for models to load
			chromedp.WaitNotPresent(`#model option`, chromedp.ByID),
			// for stability
			chromedp.Sleep(delay),
			// get the models
			chromedp.Evaluate(`Array.from(document.querySelectorAll("#model option"))
            .map(option => option.value)
            .filter(value => value !== "")`, &models),
		); err != nil {
			c.l.Error("error getting models", logger.ErrAttr(err), logger.StringAttr("brand", brand["name"]))
			// Skip this brand if there was an error or no models found
			continue
		}

		// skip if models are not found
		if len(models) == 0 {
			c.l.Warn("no models found", logger.StringAttr("brand", brand["name"]))
			continue
		}

		uniqueModels := utils.RemoveDuplicates(models)
		modelsAndBrands[brand["id"]] = uniqueModels
		c.l.Debug(fmt.Sprintf("processed brand %s, found models: %d", brand["name"], len(uniqueModels)))
	}

	c.l.Info(fmt.Sprintf("found brands: %d and successfully finished", len(modelsAndBrands)))

	return modelsAndBrands, nil
}

// GetCarChassisList retrieves the list of car body types.
//
//nolint:dupl,nolintlint
func (c *Client) GetCarChassisList(ctx context.Context) (map[string]string, error) {
	ctx, cancel := chromedp.NewRemoteAllocator(ctx, c.cfg.ChromeWSURL)
	defer cancel()

	ctxTask, taskCancel := chromedp.NewContext(ctx)
	defer taskCancel()

	c.l.Info("loading the site for car chassis list")

	if err := chromedp.Run(
		ctxTask,
		chromedp.Navigate(urlPA),
		chromedp.WaitReady(`#brand`, chromedp.ByID),
		chromedp.Sleep(delay),
	); err != nil {
		return nil, fmt.Errorf("error loading the site: %w", err)
	}

	// collect chassis types
	var chassisTypes map[string]string
	if err := chromedp.Run(
		ctxTask,
		chromedp.Evaluate(`
			(function() {
				const result = {};
				document.querySelectorAll('#chassis option').forEach(option => {
					if (option.value) {
						result[option.textContent.trim()] = option.value;
					}
				});
				return result;
			})()
		`, &chassisTypes),
	); err != nil {
		return nil, fmt.Errorf("error getting chassis types: %w", err)
	}

	c.l.Info(fmt.Sprintf("found chassis types: %d and success finished", len(chassisTypes)))

	return chassisTypes, nil
}

// GetRegionsList retrieves the list of regions.
//
//nolint:dupl,nolintlint
func (c *Client) GetRegionsList(ctx context.Context) (map[string]string, error) {
	ctx, cancel := chromedp.NewRemoteAllocator(ctx, c.cfg.ChromeWSURL)
	defer cancel()

	ctxTask, taskCancel := chromedp.NewContext(ctx)
	defer taskCancel()

	c.l.Info("loading the site for regions list")

	if err := chromedp.Run(
		ctxTask,
		chromedp.Navigate(urlPA),
		chromedp.WaitVisible("#region", chromedp.ByID), // wait for the region list to load
		chromedp.Sleep(delay),
	); err != nil {
		return nil, fmt.Errorf("error loading the site: %w", err)
	}

	// collect regions
	var regions map[string]string
	if err := chromedp.Run(
		ctxTask,
		chromedp.Evaluate(`
			(function() {
				const regions = {};
				document.querySelectorAll('#region option').forEach(option => {
					if (option.value && !/^\d+$/.test(option.value)) {
						regions[option.textContent.trim()] = option.value;
					}
				});
				return regions;
			})()
		`, &regions),
	); err != nil {
		return nil, fmt.Errorf("error getting regions: %w", err)
	}

	c.l.Info(fmt.Sprintf("found regions: %d and success finished", len(regions)))

	return regions, nil
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

	req.Header.Set("User-Agent", getRandomUserAgent())

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

// getRandomUserAgent returns a random user agent string.
func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.1355.1599 Mobile Safari/537.36",                                           //nolint:lll
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36",                                                          //nolint:lll
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",                                                              //nolint:lll
		"Mozilla/5.0 (Linux; U; Android 13; en-id; CPH2375 Build/TP1A.220905.001) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.88 Mobile Safari/537.36 HeyTapBrowser/45.10.4.1.1", //nolint:lll
		"Mozilla/5.0 (Linux; Android 14; Pixel 8 Pro Build/AP2A.240905.003; ) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.6723.107 Mobile Safari/537.36 ButtonSDK/7.0.0",             //nolint:lll
		"Mozilla/5.0 (Linux; Android 14; 2201122G Build/UKQ1.230917.001) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.6723.106 Mobile Safari/537.36 OPX/2.5",                          //nolint:lll
	}

	return userAgents[rand.IntN(len(userAgents))] //nolint:gosec,nolintlint
}
