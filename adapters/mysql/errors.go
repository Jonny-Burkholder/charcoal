package mysql

import (
	"errors"
	"fmt"
)

var (
	ErrBuild           = errors.New("unable to build MySQL query")
	ErrEmptyTable      = errors.New("config.Table must not be empty")
	ErrUnknownOperator = errors.New("unknown filter operator")
)

// UnsupportedOperatorError is returned when a filter operator constant
// has no MySQL translation.
type UnsupportedOperatorError struct {
	Operator int
}

func (e UnsupportedOperatorError) Error() string {
	return fmt.Sprintf("unsupported filter operator: %d", e.Operator)
}
