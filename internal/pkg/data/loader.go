package data

import (
	_ "embed"
	"encoding/json"

	"github.com/pkg/errors"
)

type Loader struct {
	carsList    map[string][]string
	chassisList map[string]string
	regionsList map[string]string
}

//go:embed cars/cars.json
var carsJSON []byte

//go:embed chassis/chassis.json
var chassisJSON []byte

//go:embed regions/regions.json
var regionsJSON []byte

func NewLoader() (*Loader, error) {
	loader := &Loader{
		carsList:    make(map[string][]string),
		chassisList: make(map[string]string),
		regionsList: make(map[string]string),
	}

	if err := loader.loadData(); err != nil {
		return nil, errors.Wrap(err, "failed to load data")
	}

	return loader, nil
}

// loadData loads the data from embedded JSON files into the service maps.
func (l *Loader) loadData() error {
	if err := json.Unmarshal(carsJSON, &l.carsList); err != nil {
		return errors.Wrap(err, "failed to unmarshal cars json")
	}

	if err := json.Unmarshal(chassisJSON, &l.chassisList); err != nil {
		return errors.Wrap(err, "failed to load chassis data")
	}

	if err := json.Unmarshal(regionsJSON, &l.regionsList); err != nil {
		return errors.Wrap(err, "failed to load regions data")
	}

	return nil
}

func (l *Loader) GetCarsList() map[string][]string {
	return l.carsList
}

func (l *Loader) GetChassisList() map[string]string {
	return l.chassisList
}

func (l *Loader) GetRegionsList() map[string]string {
	return l.regionsList
}
