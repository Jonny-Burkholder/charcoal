package query

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
	"strings"
)

func parseSortTokens(input string, fields filter.Fields) (tokens.SortToken, error) {
	// split by comma, then parse each segment into field and direction
	// we don't need the rigmarole of looking at quotes and parens for these,
	// those would just be invalid
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	var sortErr error

	segments := strings.Split(input, ",")
	clauses := make([]tokens.SortClause, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		if !strings.Contains(seg, ":") {
			sortErr = errors.Join(sortErr, InvalidSortExpressionError(seg))
			continue
		}

		var splitErr error

		// split on colon to separate field and direction
		parts := strings.Split(seg, ":")
		if len(parts) != 2 {
			sortErr = errors.Join(sortErr, InvalidSortExpressionError(seg))
			continue
		}

		field := strings.TrimSpace(parts[0])
		direction := strings.TrimSpace(parts[1])

		// validate field exists in field map
		if _, ok := fields[field]; !ok {
			splitErr = FieldNotFoundError(field)
		}

		// validate direction is "asc" or "desc"
		var ascending bool
		switch strings.ToLower(direction) {
		case "asc", "ascending":
			ascending = true
		case "desc", "descending":
			ascending = false
		default:
			splitErr = errors.Join(splitErr, InvalidSortDirectionError(direction))
		}

		if splitErr != nil {
			sortErr = errors.Join(sortErr, splitErr)
			continue
		}

		clauses = append(clauses, tokens.SortClause{
			Field: field,
			Asc:   ascending,
		})
	}

	if sortErr != nil {
		sortErr = errors.Join(InvalidSortExpressionError(input), sortErr)
	}

	return clauses, sortErr

}

func normalizeSortExpression(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}
