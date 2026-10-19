package query

import (
	"charcoal/internal/tokens"
	"errors"
	"strconv"
)

const maxOffsetDefault = 5000

func parsePaginationTokens(limit, offset string, maxOffset ...uint) (tokens.PaginationToken, error) {
	var token tokens.PaginationToken
	var paginationErr error

	// validate the values for each limit and offset
	if limit != "" {
		limitToken, err := strconv.Atoi(limit)
		if err != nil || limitToken < 0 {
			paginationErr = errors.Join(paginationErr, InvalidPaginationError{"limit", limit})
		} else {
			token.Limit = uint(limitToken)
		}
	}
	if offset != "" {
		maxOff := uint(maxOffsetDefault)
		if maxOffset != nil {
			maxOff = maxOffset[0]
		}
		offsetToken, err := strconv.Atoi(offset)
		if err != nil {
			paginationErr = errors.Join(paginationErr, InvalidPaginationError{"offset", offset})
		} else if offsetToken > int(maxOff) || -offsetToken > int(maxOff) {
			paginationErr = errors.Join(paginationErr, InvalidPaginationError{"offset: invalid value", offset})
		} else {
			token.Offset = offsetToken
		}
	}

	if paginationErr != nil {
		return tokens.PaginationToken{}, errors.Join(ErrInvalidPaginationExpression, paginationErr)
	}

	return token, nil
}
