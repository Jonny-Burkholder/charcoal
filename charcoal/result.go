package charcoal

import "charcoal/internal/tokens"

type Result struct {
	Tokens tokens.Tokens
	Error  error
}
