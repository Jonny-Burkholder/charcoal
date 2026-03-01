package query

import (
	"charcoal/internal/filter"
	"errors"
	"fmt"
)

var (
	ErrParsingQuery                = errors.New("unable to parse query string")
	ErrMalformedFilterString       = errors.New("malformed filter string")
	ErrMismatchedParens            = errors.New("unmatched parentheses in filter string")
	ErrMismatchedQuotes            = errors.New("unmatched quotes in filter string")
	ErrInvalidFilterExpression     = errors.New("invalid filter expression")
	ErrInvalidSortExpression       = errors.New("invalid sort expression")
	ErrInvalidPaginationExpression = errors.New("invalid pagination expression")
)

// FieldNotFoundError is returned when a field name is not present in the field map.
type FieldNotFoundError string

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf("field not found: %s", string(e))
}

// InvalidOperatorError is returned when an operator string is not recognized.
type InvalidOperatorError string

func (e InvalidOperatorError) Error() string {
	return fmt.Sprintf("invalid operator: '%s'", string(e))
}

// TypeMismatchError is returned when a value is incompatible with the field's declared type.
type TypeMismatchError struct {
	Field    string
	Value    string
	Expected filter.FieldType
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch for field '%s': value '%s' is not compatible with %s", e.Field, e.Value, fieldTypeName(e.Expected))
}

// InvalidExpressionError is returned when a filter segment cannot be parsed into field/operator/value.
type InvalidExpressionError string

func (e InvalidExpressionError) Error() string {
	return fmt.Sprintf("invalid filter expression: '%s'", string(e))
}

// InvalidSortDirectionError is returned when a sort direction is not "asc" or "desc".
type InvalidSortDirectionError string

func (e InvalidSortDirectionError) Error() string {
	return fmt.Sprintf("invalid sort direction: '%s'", string(e))
}

// InvalidPaginationError is returned when a pagination value cannot be parsed.
type InvalidPaginationError struct {
	Field string
	Value string
}

func (e InvalidPaginationError) Error() string {
	return fmt.Sprintf("invalid pagination value for '%s': '%s'", e.Field, e.Value)
}

// fieldTypeName returns a human-readable name for a FieldType constant.
func fieldTypeName(ft filter.FieldType) string {
	switch ft {
	case filter.TypeNumber:
		return "TypeNumber"
	case filter.TypeString:
		return "TypeString"
	case filter.TypeBool:
		return "TypeBool"
	case filter.TypeBytes:
		return "TypeBytes"
	default:
		return "TypeUnknown"
	}
}
