package query

import (
	"errors"
	"testing"
)

type splitFilterStringTestCase struct {
	name        string
	input       string
	expected    []string
	expectedErr error
}

var splitFilterStringTestCases = []splitFilterStringTestCase{
	{
		name:     "simple expression",
		input:    "age>30",
		expected: []string{"age>30"},
	},
	{
		name:     "expression with spaces",
		input:    "age > 30",
		expected: []string{"age > 30"},
	},
	{
		name:     "multiple expressions",
		input:    "age>30 , name='John',city='New York'",
		expected: []string{"age>30", "name='John'", "city='New York'"},
	},
	{
		name:     "expressions with commas in values",
		input:    "name='Smith, John', city='New York'",
		expected: []string{"name='Smith, John'", "city='New York'"},
	},
	{
		name:     "expressions with parentheses",
		input:    "age>30, (name='John' OR name='Jane')",
		expected: []string{"age>30", "(name='John' OR name='Jane')"},
	},
	{
		name:     "expressions with nested parentheses",
		input:    "age>30, (name='John' OR (name='Jane' AND city='NY'))",
		expected: []string{"age>30", "(name='John' OR (name='Jane' AND city='NY'))"},
	},
	{
		name:     "expressions with quoted values containing parentheses",
		input:    "name='John (Smith)', age>30",
		expected: []string{"name='John (Smith)'", "age>30"},
	},
	{
		name:     "expressions with single quote within double quotes",
		input:    `name="O'Connor", age>30`,
		expected: []string{`name="O'Connor"`, "age>30"},
	},
	{
		name:     "expressions with double quote within single quotes",
		input:    `name='She said "Hello"', age>30`,
		expected: []string{`name='She said "Hello"'`, "age>30"},
	},
	{
		name:        "unbalanced parentheses - missing close paren",
		input:       "age>30, (name='John' OR name='Jane'",
		expectedErr: ErrMismatchedParens,
	},
	{
		name:        "unbalanced parentheses - missing open paren",
		input:       "age>30, name='John' OR name='Jane')",
		expectedErr: ErrMismatchedParens,
	},
	{
		name:        "unbalanced parentheses - multiple open parens",
		input:       "age>30, (name='John' OR (name='Jane' AND city='NY')",
		expectedErr: ErrMismatchedParens,
	},
	{
		name:        "unbalanced quotes - missing closing quote",
		input:       "name='John, age>30",
		expectedErr: ErrMismatchedQuotes,
	},
	{
		name:        "unbalanced quotes - missing opening quote",
		input:       "name=John', age>30",
		expectedErr: ErrMismatchedQuotes,
	},
	{
		name:        "unbalanced quotes - multiple quotes",
		input:       "name='John, age>30, name='Jane'",
		expectedErr: ErrMismatchedQuotes,
	},
}

func TestSplitFilterString(t *testing.T) {
	for _, tc := range splitFilterStringTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := splitFilterTokens(tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error '%v', got '%v'", tc.expectedErr, err)
			}
			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tc.expected), len(result))
			}
			for i := range result {
				if result[i] != tc.expected[i] {
					t.Errorf("expected token %d to be '%s', got '%s'", i, tc.expected[i], result[i])
				}
			}
		})
	}
}

type isGroupTokenTestCase struct {
	name     string
	input    string
	expected bool
}

var isGroupTokenTestCases = []isGroupTokenTestCase{
	{
		name:     "simple group token",
		input:    "(name='John', name='Jane')",
		expected: true,
	},
	{
		name:     "nested group token",
		input:    "(name='John' OR (name='Jane' AND city='NY'))",
		expected: true,
	},
	{
		name:     "no group token",
		input:    "age>30",
		expected: false,
	},
	{
		name:     "token with quoted values",
		input:    "name='John (Smith)'",
		expected: false,
	},
	{
		name:     "malformed token - it doesn't matter",
		input:    "name='John' OR name='Jane'",
		expected: false,
	},
}

func TestIsGroupToken(t *testing.T) {
	for _, tc := range isGroupTokenTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isGroupToken(tc.input)
			if result != tc.expected {
				t.Errorf("expected '%v', got '%v'", tc.expected, result)
			}
		})
	}
}
