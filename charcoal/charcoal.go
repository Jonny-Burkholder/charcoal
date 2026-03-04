package charcoal

import (
	"charcoal/internal/filter"
	"charcoal/internal/query"
)

type charcoal struct {
	filter.Filter
	in integrations
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

func (c charcoal) Activate(queryStr string) Result {
	tokens, err := query.Parse(queryStr, c.Fields)
	return Result{
		Tokens: tokens,
		Error:  err,
	}
}
