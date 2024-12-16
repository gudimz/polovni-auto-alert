package job

import "context"

type Notifier interface {
	UpdateCarList(context.Context) error
	UpdateCarChassisList(context.Context) error
	UpdateCarRegionsList(context.Context) error
}

type Scrapper interface {
	UpdateCarChassisList(context.Context) error
}
