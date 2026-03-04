package charcoal

import (
	"charcoal/internal/filter"
	"charcoal/internal/query"
)

type charcoal struct {
	filter.Filter
}

func Filter(data any, config ...filter.Config) (charcoal, error) {
	filter, err := filter.New(data, config...)
	if err != nil {
		return charcoal{}, err
	}
	return charcoal{
		Filter: filter,
	}, nil
}

func (c charcoal) Activate(queryStr string) filter.Result {
	tokens, err := query.Parse(queryStr, c.Fields)
	return filter.Result{
		Tokens: tokens,
		Error:  err,
	}
}
