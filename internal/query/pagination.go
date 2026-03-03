package query

import (
	"charcoal/internal/tokens"
	"errors"
	"strconv"
)

func parsePaginationTokens(pagination, perPage, page, cursor string) (tokens.PaginationToken, error) {
	if pagination == "" {
		return tokens.PaginationToken{}, nil
	}

	var token tokens.PaginationToken
	var paginationErr error

	// validate the values for each pagination parameter

	paginationBool, err := strconv.ParseBool(pagination)
	if err != nil {
		paginationErr = errors.Join(paginationErr, InvalidPaginationError{"pagination", pagination})
	} else {
		token.Paginate = paginationBool
	}

	if !token.Paginate && paginationErr == nil {
		// if pagination is false, we ignore the other pagination parameters
		return token, nil
	}

	if perPage != "" {
		perPageToken, err := strconv.Atoi(perPage)
		if err != nil {
			paginationErr = errors.Join(paginationErr, InvalidPaginationError{"per_page", perPage})
		} else {
			token.PerPage = perPageToken
		}
	}
	if page != "" {
		pageToken, err := strconv.Atoi(page)
		if err != nil {
			paginationErr = errors.Join(paginationErr, InvalidPaginationError{"page", page})
		} else {
			token.Page = pageToken
		}
	}
	if cursor != "" {
		// TODO: I have to refresh on how cursor values work
		token.Cursor = cursor
	}

	if paginationErr != nil {
		return tokens.PaginationToken{}, errors.Join(ErrInvalidPaginationExpression, paginationErr)
	}

	return token, nil
}
