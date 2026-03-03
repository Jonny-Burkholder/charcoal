package query

import (
	"charcoal/internal/tokens"
	"strings"
)

// splitFilterExpression splits a single filter expression on whitespace,
// ignoring whitespace inside parentheses, brackets, or quotes
func splitFilterExpression(s string) []string {
	var segments []string
	var currentSegment []rune
	var parenDepth, bracketDepth int
	var inQuotes bool

	for _, r := range s {
		switch r {
		case '(':
			if !inQuotes {
				parenDepth++
			}
		case ')':
			if !inQuotes {
				parenDepth--
			}
		case '[':
			if !inQuotes {
				bracketDepth++
			}
		case ']':
			if !inQuotes {
				bracketDepth--
			}
		case '\'':
			inQuotes = !inQuotes
		case ' ':
			if parenDepth == 0 && bracketDepth == 0 && !inQuotes {
				if len(currentSegment) > 0 {
					segments = append(segments, string(currentSegment))
					currentSegment = []rune{}
				}
				continue
			}
		}
		currentSegment = append(currentSegment, r)
	}

	if len(currentSegment) > 0 {
		segments = append(segments, string(currentSegment))
	}

	return segments
}

// splitTopLevel splits tok on comma or "or", but only when the separator
// appears outside of single quotes, double quotes, and parentheses. The separator
// is matched case-insensitively. All resulting segments are trimmed of whitespace.
// it should be impossible for there to be mixed operators (, and "or") at this
// point. We'll cover it with extensive testing
func splitTopLevel(tok string) ([]string, tokens.JoinOp) {
	tok = strings.TrimSpace(tok)
	// remove any outer parens
	for strings.HasPrefix(tok, "(") && strings.HasSuffix(tok, ")") {
		tok = strings.TrimSpace(tok[1 : len(tok)-1])
	}

	tokLower := strings.ToLower(tok)

	var result []string
	var inSingleQuote, inDoubleQuote bool
	var parenDepth, bracketDepth int
	var op tokens.JoinOp
	lastSplit := 0

	// index all occurrences of "or" and "," in the string
	for i := 0; i < len(tok); i++ {
		switch tok[i] {
		case '\'':
			inSingleQuote = !inSingleQuote
		case '"':
			inDoubleQuote = !inDoubleQuote
		case '(':
			if !inSingleQuote && !inDoubleQuote {
				parenDepth++
			}
		case ')':
			if !inSingleQuote && !inDoubleQuote {
				parenDepth--
			}
		case '[':
			if !inSingleQuote && !inDoubleQuote {
				bracketDepth++
			}
		case ']':
			if !inSingleQuote && !inDoubleQuote {
				bracketDepth--
			}
		}

		if parenDepth == 0 && bracketDepth == 0 && !inSingleQuote && !inDoubleQuote {
			if i+1 < len(tokLower) && tokLower[i:i+2] == "or" {
				result = append(result, strings.TrimSpace(tok[lastSplit:i]))
				lastSplit = i + 2
				i++ // skip the 'r' in "or"
				op = tokens.OpOr
			} else if tok[i] == ',' {
				result = append(result, strings.TrimSpace(tok[lastSplit:i]))
				lastSplit = i + 1
				op = tokens.OpAnd
			}
		}
	}

	if lastSplit < len(tok) {
		result = append(result, strings.TrimSpace(tok[lastSplit:]))
	}

	return result, op
}
