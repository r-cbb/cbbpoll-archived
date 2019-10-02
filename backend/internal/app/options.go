package app

import (
	"github.com/r-cbb/cbbpoll/internal/db"
	"time"
)

type Options struct {
	filters []db.Filter
	sort db.Sort
}

func NewOptions() Options {
	opt := Options{
		filters: make([]db.Filter, 0),
	}

	return opt
}

func (opt Options) unpack() ([]db.Filter, db.Sort) {
	return opt.filters, opt.sort
}

func (opt Options) IsVoter(b bool) Options {
	opt.filters = append(opt.filters, db.Filter{Field: "IsVoter", Operator: "=", Value: b})
	return opt
}

func (opt Options) HasOpened() Options {
	opt.filters = append(opt.filters, db.Filter{Field: "OpenTime", Operator: "<", Value: time.Now()})
	return opt
}