package fetcher

import (
	"context"
)

type (
	PolovniAutoAdapter interface {
		GetCarsList(context.Context) (map[string][]string, error)
		GetCarChassisList(context.Context) (map[string]string, error)
		GetRegionsList(context.Context) (map[string]string, error)
	}
)
