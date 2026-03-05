package charcoal

import (
	"charcoal/internal/filter"
	"charcoal/internal/query"
	"charcoal/internal/tokens"
)

type charcoal struct {
	filter.Filter
	builtin
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

func (c charcoal) Activate(queryStr string) (tokens.Tokens, error) {
	return query.Parse(queryStr, c.Fields)
}
