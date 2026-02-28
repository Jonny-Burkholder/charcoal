package query

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
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

// parseFilterToken takes a filter query substring and parses the field, operator, and value into
// a FilterToken. It returns an error if the token is malformed.
func parseFilterToken(tok string, fields filter.Fields) (tokens.FilterToken, error) {
	token := tokens.FilterToken{}
	clause := tokens.Clause{}

	tok = strings.TrimSpace(tok)

	var op *int // to check zero value
	var field string
	var fieldType filter.FieldType

	// first, split on whitespace and see if we can get the whole token
	parts := strings.Fields(tok)
	if len(parts) >= 2 {
		field = parts[0]
		if typ, ok := fields[field]; ok {
			fieldType = typ
		}
		opCandidate := parts[1]
		if opInt, ok := filter.OperatorMap[opCandidate]; ok {
			op = &opInt
		}
	}

	if field != "" && op != nil {
		clause.Field = field
		clause.Operator = *op
	}

	// if the operator is null or not null, we don't need a value.
	// but if there is a value, the token is malformed

	// if the operator is null or not null, we need to fetch a value
	// if there's too many or too few parts, the token is malformed

	// if that didn't work, search for canonical operators in the token not within quotes
	var inSingleQuote, inDoubleQuote bool
	for i := 0; i < len(tok); i++ {
		r := tok[i]
		switch r {
		case runeSingleQuote:
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case runeDoubleQuote:
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		default:
			if !inSingleQuote && !inDoubleQuote {
				// check for operators at this position
				for opStr := range filter.OperatorMap {
					if strings.HasPrefix(tok[i:], opStr) {
						op = opStr
						opIndex = i
						break
					}
				}
				if op != "" {
					break
				}
			}
		}
	}

	return tokens.FilterToken{}, nil
}

// parseGroupToken takes a group query substring and parses it into a Token tree
// It returns an error if the token is malformed.
func parseGroupToken(tok string) ([]tokens.FilterToken, error) {
	return []tokens.FilterToken{}, nil
}
