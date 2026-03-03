package query

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// some runes for reference
const (
	runeSpace        = ' '
	runeComma        = ','
	runeSingleQuote  = '\''
	runeDoubleQuote  = '"'
	runeOpenParen    = '('
	runeCloseParen   = ')'
	runeOpenBracket  = '['
	runeCloseBracket = ']'
)

// Parse parses a URL-encoded query string into a db-agnostic token tree.
// The query string should contain URL query parameters such as filter, sort,
// pagination, per_page, page, and cursor.
func Parse(queryStr string, fields filter.Fields) (tokens.Tokens, error) {
	toks := tokens.Tokens{}
	var parseErr error

	// parse URL query parameters
	params, err := url.ParseQuery(queryStr)
	if err != nil {
		return tokens.Tokens{}, errors.Join(ErrParsingQuery, err)
	}

	filterStr := params.Get("filters")
	sortStr := params.Get("sort")
	paginationStr := params.Get("pagination")
	perPageStr := params.Get("per_page")
	pageStr := params.Get("page")
	cursorStr := params.Get("cursor")

	var filterTokens []tokens.FilterToken
	var clauses []tokens.Clause
	// parse filter tokens
	if filterStr != "" {
		expressions, err := splitFilterTokens(filterStr)
		if err != nil {
			parseErr = errors.Join(parseErr, err)
		} else {
			for _, expr := range expressions {
				if isGroupToken(expr) {
					ft, err := parseGroupToken(expr, fields)
					if err != nil {
						parseErr = errors.Join(parseErr, err)
						continue
					}
					filterTokens = append(filterTokens, ft)
				} else {
					clause, err := parseFilterClause(expr, fields)
					if err != nil {
						parseErr = errors.Join(parseErr, err)
						continue
					}
					clauses = append(clauses, clause)
				}
			}
		}
	}

	filterTokens = append(filterTokens, tokens.FilterToken{
		Clauses: clauses,
	})

	toks.Filter = filterTokens

	// parse sort tokens
	if sortStr != "" {
		sortTokens, err := parseSortTokens(sortStr, fields)
		if err != nil {
			parseErr = errors.Join(parseErr, err)
		} else {
			toks.Sort = sortTokens
		}
	}

	// parse pagination tokens
	if paginationStr != "" {
		paginationToken, err := parsePaginationTokens(paginationStr, perPageStr, pageStr, cursorStr)
		if err != nil {
			parseErr = errors.Join(parseErr, err)
		} else {
			toks.Pagination = paginationToken
		}
	}

	if parseErr != nil {
		return tokens.Tokens{}, errors.Join(ErrParsingQuery, parseErr)
	}

	return toks, nil
}

// splitFilterTokens splits a filter string into its component expressions, respecting parentheses
// and quoted values. It returns an error if quotes or parentheses are unbalanced.
func splitFilterTokens(queryStr string) ([]string, error) {
	var res []string
	var splitErr error

	// remove trailing commas and whitespace
	queryStr = strings.Trim(queryStr, " ,")

	// index all commas that are not within parentheses or quotes
	// if there's a single or double quote, we can ignore the rest
	// until we find the matching quote
	// if there's an open paren, we can ignore the rest until we find the matching close paren
	// if we find a comma and we're not within quotes or parentheses, we split there
	var indexes []int
	var inSingleQuote, inDoubleQuote, inParens, inBrackets bool
	var parenDepth, bracketDepth int
	for i, r := range queryStr {
		switch r {
		case runeSingleQuote:
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case runeDoubleQuote:
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case runeOpenParen:
			if !inSingleQuote && !inDoubleQuote && !inBrackets {
				inParens = true
				parenDepth++
			}
		case runeCloseParen:
			if !inSingleQuote && !inDoubleQuote && !inBrackets {
				parenDepth--
				if parenDepth == 0 {
					inParens = false
				} else if parenDepth < 0 {
					splitErr = errors.Join(splitErr, ErrMismatchedParens)
				}
			}
		case runeOpenBracket:
			if !inSingleQuote && !inDoubleQuote {
				inBrackets = true
				bracketDepth++
			}
		case runeCloseBracket:
			if !inSingleQuote && !inDoubleQuote {
				bracketDepth--
				if bracketDepth == 0 {
					inBrackets = false
				} else if bracketDepth < 0 {
					splitErr = errors.Join(splitErr, ErrMismatchedBrackets)
				}
			}
		case runeComma:
			if !inSingleQuote && !inDoubleQuote && !inParens && !inBrackets {
				indexes = append(indexes, i)
			}
		}
	}

	if inSingleQuote || inDoubleQuote {
		splitErr = errors.Join(splitErr, ErrMismatchedQuotes)
	}

	if parenDepth != 0 || inParens {
		splitErr = errors.Join(splitErr, ErrMismatchedParens)
	}

	if bracketDepth != 0 || inBrackets {
		splitErr = errors.Join(splitErr, ErrMismatchedBrackets)
	}

	if splitErr != nil {
		return nil, splitErr
	}

	if len(indexes) < 1 {
		// there's only one clause, return it
		return []string{strings.TrimSpace(queryStr)}, nil
	}

	// if there's only one comma, return both parts
	if len(indexes) == 1 {
		return []string{
			strings.TrimSpace(queryStr[:indexes[0]]),
			strings.TrimSpace(queryStr[indexes[0]+1:]),
		}, nil
	}

	// get the first part before the first comma
	res = append(res, strings.TrimSpace(queryStr[:indexes[0]]))

	// split the string at the identified indexes
	for i := 0; i < len(indexes)-1; i++ {
		res = append(res, strings.TrimSpace(queryStr[indexes[i]+1:indexes[i+1]]))
	}
	// add the final segment after the last comma
	res = append(res, strings.TrimSpace(queryStr[indexes[len(indexes)-1]+1:]))

	return res, nil
}

// isGroupToken returns true under one of two conditions:
// 1. The token contains unquoted parentheses with valid separators (commas or "OR") within them.
// 2. The token contains unquoted "OR" operators outside of any parentheses.
func isGroupToken(tok string) bool {
	var inSingleQuote, inDoubleQuote, inParen, inBracket bool
	var parenDepth, bracketDepth int
	tok = strings.ToLower(tok)

	for i, r := range tok {
		switch r {
		case runeSingleQuote:
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case runeDoubleQuote:
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case runeOpenBracket:
			if !inSingleQuote && !inDoubleQuote {
				inBracket = true
				bracketDepth++
			}
		case runeCloseBracket:
			if !inSingleQuote && !inDoubleQuote {
				bracketDepth--
				if bracketDepth == 0 {
					inBracket = false
				}
			}
		case runeOpenParen:
			if !inSingleQuote && !inDoubleQuote && !inBracket {
				inParen = true
				parenDepth++
			}
		case runeCloseParen:
			if !inSingleQuote && !inDoubleQuote && !inBracket {
				parenDepth--
				if parenDepth == 0 {
					inParen = false
				}
			}
		case runeComma:
			if !inSingleQuote && !inDoubleQuote && !inBracket && inParen {
				return true
			}
		case 'o':
			if !inSingleQuote && !inDoubleQuote && !inBracket {
				// check if the next characters are "or" with word boundaries
				if i+1 < len(tok) && tok[i+1] == 'r' {
					prevOk := i == 0 || tok[i-1] == ' '
					nextOk := i+2 >= len(tok) || tok[i+2] == ' '
					if prevOk && nextOk {
						return true
					}
				}
			}
		case 'a':
			if !inSingleQuote && !inDoubleQuote && !inBracket {
				// check if the next characters are "and" with word boundaries
				if i+2 < len(tok) && tok[i+1] == 'n' && tok[i+2] == 'd' {
					prevOk := i == 0 || tok[i-1] == ' '
					nextOk := i+3 >= len(tok) || tok[i+3] == ' '
					if prevOk && nextOk {
						return true
					}
				}
			}
		}
	}

	return false
}

// TODO: this function can be more DRY. A LOT more
// parseFilterClause takes a filter query substring and parses the field, operator, and value into
// a tokens.Clause. It returns an error if the clause is malformed.
func parseFilterClause(tok string, fields filter.Fields) (tokens.Clause, error) {
	clause := tokens.Clause{}

	// normalize the token string
	tok = strings.TrimSpace(tok)

	var op *int // to check zero value
	var field string
	var fieldType filter.FieldType
	var splitOp bool // to track if the user submitted a two-word operator like "not in"
	var found bool   // did we find the token with this string split method?

	// first, split on whitespace and see if we can get the whole token
	parts := splitFilterExpression(tok)

	var parseError error

	// we'll just assume if there's the correct number of parts, the
	// user is attempting to split on whitespace
	if len(parts) >= 2 {
		field = normalizeCandidate(parts[0])
		if typ, ok := fields[field]; ok {
			fieldType = typ
		} else {
			parseError = errors.Join(parseError, FieldNotFoundError(field))
		}
		opCandidate := normalizeCandidate(parts[1])
		if opInt, ok := filter.OperatorMap[opCandidate]; ok {
			op = &opInt
		}
		// if that doesn't work, see if the op candidate is the first part of
		// a valid operator
		if op == nil && slices.Contains(filter.FirstWords, opCandidate) {
			// get the second part and see if the two together are a valid operator
			if len(parts) >= 3 {
				opCandidate = opCandidate + " " + parts[2]
				if opInt, ok := filter.OperatorMap[opCandidate]; ok {
					op = &opInt
					splitOp = true
				}
			}
		}
		if op == nil {
			if len(parts) == 3 {
				opCandidate = normalizeCandidate(parts[1])
			}
			parseError = errors.Join(parseError, InvalidOperatorError(opCandidate))
		}
	}

	if parseError != nil {
		return tokens.Clause{}, errors.Join(ErrInvalidFilterExpression, parseError)
	}

	if field != "" && op != nil {
		clause.Field = field
		clause.Operator = *op
		found = true
	}

	// logic to do if we found an operator with the above method
	if found {
		// if the operator is null or not null, we don't need a value.
		// but if there is a value, the token is malformed
		if *op == filter.OpIsNull || *op == filter.OpNotNull {
			if (!splitOp && len(parts) > 2) || (splitOp && len(parts) > 3) {
				return tokens.Clause{}, InvalidExpressionError(tok)
			}
			return clause, nil
		}

		// for "in" and "not in" operators, the expected value is a bracket-delimited list.
		if *op == filter.OpIn || *op == filter.OpNotIn {
			valueStart := 2
			if splitOp {
				valueStart = 3
			}

			if len(parts) != valueStart+1 {
				return tokens.Clause{}, InvalidExpressionError(tok)
			}

			value := parts[valueStart]
			if !validateSliceType(value, fieldType) {
				return tokens.Clause{}, TypeMismatchError{
					Field:    field,
					Value:    value,
					Expected: fieldType,
				}
			}

			value = normalizeFieldValue(value)
			if !fieldTypeIsValid(value, fieldType) {
				return tokens.Clause{}, TypeMismatchError{
					Field:    field,
					Value:    value,
					Expected: fieldType,
				}
			}

			clause.Value = value
			return clause, nil
		}

		// otherwise, we need to fetch a value
		// if there's too many or too few parts, the token is malformed
		if (!splitOp && len(parts) != 3) || (splitOp && len(parts) != 4) {
			return tokens.Clause{}, InvalidExpressionError(tok)
		}

		// the value is either the third part or the fourth part, depending on whether the operator was split
		value := parts[len(parts)-1]

		// normalize the filter value and validate that the value is compatible with the field type
		value = normalizeFieldValue(value)
		if !fieldTypeIsValid(value, fieldType) {
			return tokens.Clause{}, TypeMismatchError{
				Field:    field,
				Value:    value,
				Expected: fieldType,
			}
		}
		clause.Value = value

		return clause, nil
	}

	// if that didn't work - indexing!
	// index the first quote instance - we'll only look before this
	// If the operator is after the first quote, get good newb
	quoteIndex := strings.IndexAny(tok, `'"`)
	subString := normalizeCandidate(tok)
	if quoteIndex != -1 {
		subString = tok[:quoteIndex]
	}

	var opString string
	var opIndex int

	// index any operator in the substring before the first quote
	for opCandidate := range filter.OperatorMap {
		if idx := strings.Index(subString, opCandidate); idx != -1 {
			opString = opCandidate
			opIndex = idx
		}
	}

	if opString == "" {
		return tokens.Clause{}, InvalidExpressionError(tok)
	}

	// we've got the operator and its index, let's split the string into field and value parts
	field = normalizeCandidate(tok[:opIndex])
	value := normalizeFieldValue(tok[opIndex+len(opString):])

	// validate the field exists in the field map
	if typ, ok := fields[field]; ok {
		fieldType = typ
	} else {
		return tokens.Clause{}, FieldNotFoundError(field)
	}

	// validate the operator
	if opInt, ok := filter.OperatorMap[opString]; ok {
		op = &opInt
	} else {
		return tokens.Clause{}, InvalidOperatorError(opString)
	}

	// validate that the value is compatible with the field type
	if *op == filter.OpIn || *op == filter.OpNotIn {
		if !validateSliceType(value, fieldType) {
			return tokens.Clause{}, TypeMismatchError{
				Field:    field,
				Value:    value,
				Expected: fieldType,
			}
		}
	} else {
		if !fieldTypeIsValid(value, fieldType) {
			return tokens.Clause{}, TypeMismatchError{
				Field:    field,
				Value:    value,
				Expected: fieldType,
			}
		}
	}

	clause.Field = field
	clause.Operator = *op
	clause.Value = value

	return clause, nil
}

// TODO: maybe return the indexes of the invalid values
func validateSliceType(value string, fieldType filter.FieldType) bool {
	// remove surrounding brackets and split on commas
	val := strings.Trim(value, "[] ")
	elements := strings.SplitSeq(val, ",")

	for elem := range elements {
		elem = strings.TrimSpace(elem)
		if !fieldTypeIsValid(elem, fieldType) {
			return false
		}
	}

	return true
}

// stripOuterParens removes the outermost parentheses from a token string if they
// wrap the entire expression. It does not strip parens in cases like "(a) OR (b)"
// where the first open paren closes before the end.
func stripOuterParens(tok string) string {
	tok = strings.TrimSpace(tok)
	if len(tok) < 2 || tok[0] != '(' {
		return tok
	}

	// find the matching close paren for the opening paren at index 0
	depth := 0
	var inSingleQuote, inDoubleQuote bool
	for i := 0; i < len(tok); i++ {
		ch := tok[i]
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		} else if !inSingleQuote && !inDoubleQuote {
			switch ch {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					if i == len(tok)-1 {
						// the matching close paren is the last char — safe to strip
						return strings.TrimSpace(tok[1 : len(tok)-1])
					}
					// the matching close paren is not at the end, don't strip
					return tok
				}
			}
		}
	}
	return tok
}

// parseGroupToken takes a group query substring and parses it into a Token tree.
// It recursively splits the token into its component expressions and parses each one.
// It returns an error if the token is malformed.
// TODO: what we'll do to clean this up is 1. validate that the token is correct (mostly
// that it only has one type of separator at the top level) 2. split based on the separator
// that we got during validation. That will clean things up a lot
func parseGroupToken(tok string, fields filter.Fields) (tokens.FilterToken, error) {
	tok = strings.TrimSpace(tok)
	tok = stripOuterParens(tok)

	result := tokens.FilterToken{}

	// split the top level
	segments, op := splitTopLevel(tok)
	result.JoinOp = op

	var groupErr error

	for _, segment := range segments {
		if isGroupToken(segment) {
			child, err := parseGroupToken(segment, fields)
			if err != nil {
				groupErr = errors.Join(groupErr, err)
				continue
			}
			nextToken := tokens.NextFilterToken{
				Op: op,
				T:  child,
			}
			result.Children = append(result.Children, nextToken)
		} else {
			clause, err := parseFilterClause(segment, fields)
			if err != nil {
				groupErr = errors.Join(groupErr, err)
				continue
			}
			result.Clauses = append(result.Clauses, clause)
		}
	}

	if groupErr != nil {
		return tokens.FilterToken{}, errors.Join(ErrInvalidFilterExpression, groupErr)
	}

	return result, nil
}

// normalizeCandidate normalizes a candidate field or operator string by trimming whitespace and converting to lowercase.
func normalizeCandidate(candidate string) string {
	return strings.ToLower(strings.TrimSpace(candidate))
}

// normalizeFieldValue removes surrounding quotes from a field value and unescapes any escaped quotes within the value.
func normalizeFieldValue(value string) string {
	value = strings.TrimSpace(value)

	// Only remove outermost quotes if they match
	if len(value) >= 2 {
		firstChar := value[0]
		lastChar := value[len(value)-1]
		if (firstChar == '"' && lastChar == '"') || (firstChar == '\'' && lastChar == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\'`, `'`)

	// final whitespace trim after unescaping
	return strings.TrimSpace(value)
}

// fieldTypeIsValid checks if the value is compatible with the field type.
// For bracket-delimited list values like [a, b, c], each element is validated individually.
func fieldTypeIsValid(value string, fieldType filter.FieldType) bool {
	// check for bracket-delimited list values (used with "in" / "not in")
	if len(value) >= 2 && value[0] == '[' && value[len(value)-1] == ']' {
		inner := value[1 : len(value)-1]
		elements := strings.Split(inner, ",")
		for _, elem := range elements {
			elem = strings.TrimSpace(elem)
			if elem == "" {
				continue
			}
			if !fieldTypeIsValid(elem, fieldType) {
				return false
			}
		}
		return true
	}

	switch fieldType {
	case filter.TypeString:
		return true // any value can be a string
	case filter.TypeNumber:
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return true
		}
	case filter.TypeBool:
		if _, err := strconv.ParseBool(value); err == nil {
			return true
		}
	case filter.TypeBytes:
		// idk what to do with this tbh
		return true
	}
	return false
}
