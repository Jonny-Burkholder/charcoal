package query

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// some runes for reference
const (
	runeSpace       = ' '
	runeComma       = ','
	runeSingleQuote = '\''
	runeDoubleQuote = '"'
	runeOpenParen   = '('
	runeCloseParen  = ')'
)

// Parse parses a query string into a db-agnostic token tree
func Parse(queryStr string, fields filter.Fields) (tokens.Tokens, error) {
	toks := tokens.Tokens{}
	var parseErr error

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
	var inSingleQuote, inDoubleQuote, inParens bool
	var parenDepth int
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
			if !inSingleQuote && !inDoubleQuote {
				inParens = true
				parenDepth++
			}
		case runeCloseParen:
			if !inSingleQuote && !inDoubleQuote {
				parenDepth--
				if parenDepth == 0 {
					inParens = false
				} else if parenDepth < 0 {
					splitErr = errors.Join(splitErr, ErrMismatchedParens)
				}
			}
		case runeComma:
			if !inSingleQuote && !inDoubleQuote && !inParens {
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

	fmt.Println(res)

	return res, nil
}

// isGroupToken returns true under one of two conditions:
// 1. The token contains unquoted parentheses with valid separators (commas or "OR") within them.
// 2. The token contains unquoted "OR" operators outside of any parentheses.
func isGroupToken(tok string) bool {
	var inSingleQuote, inDoubleQuote, inParen bool
	var parenDepth int
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
		case runeOpenParen:
			if !inSingleQuote && !inDoubleQuote {
				inParen = true
				parenDepth++
			}
		case runeCloseParen:
			if !inSingleQuote && !inDoubleQuote {
				parenDepth--
				if parenDepth == 0 {
					inParen = false
				}
			}
		case runeComma:
			if !inSingleQuote && !inDoubleQuote && inParen {
				return true
			}
		case 'o':
			if !inSingleQuote && !inDoubleQuote {
				// check if the next characters are "or"
				if i+1 < len(tok) && tok[i+1] == 'r' {
					return true
				}
			}
		}
	}

	return false
}

// TODO: this function can be more DRY
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
	parts := strings.Fields(tok)
	if len(parts) >= 2 {
		field = normalizeCandidate(parts[0])
		if typ, ok := fields[field]; ok {
			fieldType = typ
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
	if !fieldTypeIsValid(value, fieldType) {
		return tokens.Clause{}, TypeMismatchError{
			Field:    field,
			Value:    value,
			Expected: fieldType,
		}
	}

	clause.Field = field
	clause.Operator = *op
	clause.Value = value

	return clause, nil
}

// parseGroupToken takes a group query substring and parses it into a Token tree
// It returns an error if the token is malformed.
func parseGroupToken(tok string) ([]tokens.FilterToken, error) {
	// TODO: logic
	return []tokens.FilterToken{}, nil
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

// fieldTypeIsValid checks if the value is compatible with the field type
func fieldTypeIsValid(value string, fieldType filter.FieldType) bool {
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
