package filter

import "charcoal/internal/tokens"

type Result struct {
	Tokens tokens.Tokens
	Error  error
}
