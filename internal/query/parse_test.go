package query

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
	"testing"
)

type splitFilterTokensTestCase struct {
	name        string
	input       string
	expected    []string
	expectedErr error
}

var splitFilterTokensTestCases = []splitFilterTokensTestCase{
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
		name:  "even more expressions",
		input: "age>30 , name='John', city='New York', isActive=true, description='A user from New York'",
		expected: []string{
			"age>30",
			"name='John'",
			"city='New York'",
			"isActive=true",
			"description='A user from New York'",
		},
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

func TestSplitFilterTokens(t *testing.T) {
	for _, tc := range splitFilterTokensTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := splitFilterTokens(tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error '%v', got '%v'", tc.expectedErr, err)
			}
			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d expressions, got %d", len(tc.expected), len(result))
			}
			for i := range result {
				if result[i] != tc.expected[i] {
					t.Errorf("expected expression %d to be '%s', got '%s'", i, tc.expected[i], result[i])
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
		expected: true,
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

type parseFilterClauseTestCase struct {
	name        string
	input       string
	fields      filter.Fields
	expected    tokens.Clause
	expectedErr error
}

var parseFilterClauseTestCases = []parseFilterClauseTestCase{
	{
		name:  "simple expression",
		input: "age>30",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: tokens.Clause{
			Field:    "age",
			Operator: filter.OpGt,
			Value:    "30",
		},
	},
	{
		name:  "simple expression with spaces",
		input: "age > 30",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: tokens.Clause{
			Field:    "age",
			Operator: filter.OpGt,
			Value:    "30",
		},
	},
	{
		name:  "simple expression with string operator",
		input: "age gt 30",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: tokens.Clause{
			Field:    "age",
			Operator: filter.OpGt,
			Value:    "30",
		},
	},
	{
		name:  "string expression with quotes",
		input: "description='John is > 30 years old'",
		fields: filter.Fields{
			"description": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "description",
			Operator: filter.OpEq,
			Value:    "John is > 30 years old",
		},
	},
	{
		name:  "string expression with double quotes and single quote inside",
		input: `description="John's book"`,
		fields: filter.Fields{
			"description": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "description",
			Operator: filter.OpEq,
			Value:    "John's book",
		},
	},
	{
		name:  "string expression with single quotes and double quote inside",
		input: `description='She said "Hello"'`,
		fields: filter.Fields{
			"description": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "description",
			Operator: filter.OpEq,
			Value:    `She said "Hello"`,
		},
	},
	{
		name:  "not null operator",
		input: "name not null	",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "name",
			Operator: filter.OpNotNull,
			Value:    "",
		},
	},
	{
		name:  "null operator",
		input: "name null",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "name",
			Operator: filter.OpIsNull,
		},
	},

	{
		name:  "is null operator",
		input: "name is null",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "name",
			Operator: filter.OpIsNull,
		},
	},
	// TODO: can this case be made valid with our current parsing logic?
	// {
	// 	name:        "invalid operator",
	// 	input:       "age >> 30",
	// 	fields:      filter.Fields{"age": filter.TypeNumber},
	// 	expectedErr: InvalidOperatorError(">>"),
	// },
	{
		name:  "invalid operator causes invalid value error",
		input: "age>>30",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: TypeMismatchError{
			Field:    "age",
			Value:    ">30",
			Expected: filter.TypeNumber,
		},
	},
	{
		name:        "field not found",
		input:       "height>180",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expectedErr: FieldNotFoundError("height"),
	},
	{
		name:   "type mismatch - non-numeric value for number field",
		input:  "age>old",
		fields: filter.Fields{"age": filter.TypeNumber},
		expectedErr: TypeMismatchError{
			Field:    "age",
			Value:    "old",
			Expected: filter.TypeNumber,
		},
	},
	{
		name:  "string number works for number field",
		input: "age>'30'",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: tokens.Clause{
			Field:    "age",
			Operator: filter.OpGt,
			Value:    "30",
		},
	},
	{
		name:  "string with spaces works for number field",
		input: "age>' 30 '",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: tokens.Clause{
			Field:    "age",
			Operator: filter.OpGt,
			Value:    "30",
		},
	},
	{
		name:  "type mismatch - non-boolean value for boolean field",
		input: "isActive>yes",
		fields: filter.Fields{
			"isactive": filter.TypeBool,
		},
		expectedErr: TypeMismatchError{
			Field:    "isactive",
			Value:    "yes",
			Expected: filter.TypeBool,
		},
	},
	{
		name:  "boolean true value",
		input: "isActive=true",
		fields: filter.Fields{
			"isactive": filter.TypeBool,
		},
		expected: tokens.Clause{
			Field:    "isactive",
			Operator: filter.OpEq,
			Value:    "true",
		},
	},
	{
		name:  "not enough parts in expression",
		input: "age30",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: InvalidExpressionError("age30"),
	},
	{
		name:  "too many parts in expression for non-split operator",
		input: "age > 30 extra",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: InvalidExpressionError("age > 30 extra"),
	},
	{
		name:  "too many parts in expression for split operator",
		input: "name not null extra",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expectedErr: InvalidExpressionError("name not null extra"),
	},
	{
		name:  "too many operators in expression",
		input: "age>30>20",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: TypeMismatchError{
			Field:    "age",
			Value:    "30>20",
			Expected: filter.TypeNumber,
		},
	},
	{
		name:  "invalid expression - no operator",
		input: "age30",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: InvalidExpressionError("age30"),
	},
}

func TestParseFilterClause(t *testing.T) {
	for _, tc := range parseFilterClauseTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseFilterClause(tc.input, tc.fields)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error '%v', got '%v'", tc.expectedErr, err)
			}
			if !clausesAreEqual(result, tc.expected) {
				t.Errorf("expected token '%v', got '%v'", tc.expected, result)
			}
		})
	}
}

func clausesAreEqual(a, b tokens.Clause) bool {
	return a.Field == b.Field && a.Operator == b.Operator && a.Value == b.Value
}

func tokensAreEqual(a, b tokens.FilterToken) bool {
	if len(a.Clauses) != len(b.Clauses) {
		return false
	}
	for i := range a.Clauses {
		if a.Clauses[i] != b.Clauses[i] {
			return false
		}
	}
	if a.JoinOp != b.JoinOp {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if a.Children[i].Op != b.Children[i].Op || !tokensAreEqual(a.Children[i].T, b.Children[i].T) {
			return false
		}
	}
	return true
}
