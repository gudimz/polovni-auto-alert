package job

import "context"

type Notifier interface {
	UpdateCarList(context.Context) error
	UpdateChassisList(context.Context) error
	UpdateRegionsList(context.Context) error
}

type Scrapper interface {
	UpdateCarChassisList(context.Context) error
}
