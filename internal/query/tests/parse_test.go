package tests

import (
	"charcoal/internal/tokens"
	"testing"
)

type parseTestCase struct {
	name        string
	query       string
	expected    []tokens.Tokens
	expectError error
}

var parseTestCases = []parseTestCase{}

func TestParse(t *testing.T) {}
