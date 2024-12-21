package fetcher

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/gudimz/polovni-auto-alert/pkg/logger"
)

type Service struct {
	l         *logger.Logger
	paAdapter PolovniAutoAdapter
}

const countGoroutines = 3

//go:embed data/regions.json
var regionsJSON []byte

//go:embed data/chassis.json
var chassisJSON []byte

//go:embed data/cars.json
var carsJSON []byte

const (
	regionsPath = "data/regions.json"
	chassisPath = "data/chassis.json"
	carsPath    = "data/cars.json"
)

// NewService creates a new Fetcher Service instance.
func NewService(l *logger.Logger, paAdapter PolovniAutoAdapter) *Service {
	return &Service{
		l:         l,
		paAdapter: paAdapter,
	}
}

// Start begins the fetching process.
func (s *Service) Start(ctx context.Context) error {
	return s.Fetch(ctx)
}

// Fetch fetches regions, chassis and cars.
func (s *Service) Fetch(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(countGoroutines)

	errChan := make(chan error, countGoroutines)

	go func() {
		defer wg.Done()

		if err := s.fetchRegions(ctx); err != nil {
			s.l.Error("failed to fetch regions", logger.ErrAttr(err))
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()

		if err := s.fetchChassis(ctx); err != nil {
			s.l.Error("failed to fetch chassis", logger.ErrAttr(err))
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()

		if err := s.fetchCars(ctx); err != nil {
			s.l.Error("failed to fetch cars", logger.ErrAttr(err))
			errChan <- err
		}
	}()

	wg.Wait()

	s.l.Info("fetching completed")
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// GetRegionsFromJSON returns regions from the regions.json.
func (s *Service) GetRegionsFromJSON() (map[string]string, error) {
	regions := make(map[string]string)
	if err := json.Unmarshal(regionsJSON, &regions); err != nil {
		return nil, err
	}

	return regions, nil
}

// GetChassisFromJSON returns car body types from the chassis.json.
func (s *Service) GetChassisFromJSON() (map[string]string, error) {
	chassis := make(map[string]string)
	if err := json.Unmarshal(chassisJSON, &chassis); err != nil {
		return nil, err
	}

	return chassis, nil
}

// GetCarsFromJSON returns car brands and models from the cars.json.
func (s *Service) GetCarsFromJSON() (map[string][]string, error) {
	cars := make(map[string][]string)
	if err := json.Unmarshal(carsJSON, &cars); err != nil {
		return nil, err
	}

	return cars, nil
}

// fetchRegions fetches regions from the https://www.polovniautomobili.com.
func (s *Service) fetchRegions(ctx context.Context) error {
	regionList, err := s.paAdapter.GetRegionsList(ctx)
	if err != nil {
		s.l.Error("failed to get regions", logger.ErrAttr(err))
		return err
	}

	if err = s.saveToFile(regionsPath, regionList); err != nil {
		s.l.Error("failed to save regions to file", logger.ErrAttr(err))
		return err
	}

	return nil
}

// fetchChassis fetches car body types from the https://www.polovniautomobili.com.
func (s *Service) fetchChassis(ctx context.Context) error {
	chassisList, err := s.paAdapter.GetCarChassisList(ctx)
	if err != nil {
		s.l.Error("failed to get chassis", logger.ErrAttr(err))
		return err
	}

	if err = s.saveToFile(chassisPath, chassisList); err != nil {
		s.l.Error("failed to save chassis to file", logger.ErrAttr(err))
		return err
	}

	return nil
}

// fetchCars fetches car brands and models from the https://www.polovniautomobili.com.
func (s *Service) fetchCars(ctx context.Context) error {
	carsList, err := s.paAdapter.GetCarsList(ctx)
	if err != nil {
		s.l.Error("failed to get cars list", logger.ErrAttr(err))
		return err
	}

	if err = s.saveToFile(carsPath, carsList); err != nil {
		s.l.Error("failed to save cars list to file", logger.ErrAttr(err))
		return err
	}

	return nil
}

// saveToFile saves data to a file.
func (s *Service) saveToFile(filename string, data any) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // pretty print

	return encoder.Encode(data)
}
