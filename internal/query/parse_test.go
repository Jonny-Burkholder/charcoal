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
	{
		name:     "expression with bracket list value",
		input:    "name in [john, barb, keith], age > 30",
		expected: []string{"name in [john, barb, keith]", "age > 30"},
	},
	{
		name:     "expression with bracket list and other filters",
		input:    "city = 'NY', name in [john, barb, keith], age > 30",
		expected: []string{"city = 'NY'", "name in [john, barb, keith]", "age > 30"},
	},
	{
		name:     "bracket list with spaces",
		input:    "name in [john, barb, keith]",
		expected: []string{"name in [john, barb, keith]"},
	},
	{
		name:        "unbalanced brackets - missing close bracket",
		input:       "name in [john, barb, keith",
		expectedErr: ErrMismatchedBrackets,
	},
	{
		name:        "unbalanced brackets - missing open bracket",
		input:       "name in john, barb, keith]",
		expectedErr: ErrMismatchedBrackets,
	},
	{
		name:  "the one that's failing the query test",
		input: "age>>30, narm = 'John' OR name is 'Jane'",
		expected: []string{
			"age>>30",
			"narm = 'John' OR name is 'Jane'",
		},
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
		name:     "no parens - it doesn't matter",
		input:    "name='John' OR name='Jane'",
		expected: true,
	},
	{
		name:     "in query with brackets is not a group token",
		input:    "name in [john, barb, keith]",
		expected: false,
	},
	{
		name:     "not in query with brackets is not a group token",
		input:    "name not in [john, barb, keith]",
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
	{
		name:        "invalid operator",
		input:       "age >> 30",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expectedErr: InvalidOperatorError(">>"),
	},
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
	{
		name:  "in operator with bracket list",
		input: "name in [john, barb, keith]",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "name",
			Operator: filter.OpIn,
			Value:    "[john, barb, keith]",
		},
	},
	{
		name:  "not in operator with bracket list",
		input: "name not in [john, barb, keith]",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expected: tokens.Clause{
			Field:    "name",
			Operator: filter.OpNotIn,
			Value:    "[john, barb, keith]",
		},
	},
	{
		name:  "in operator with number list",
		input: "age in [20, 30, 40]",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: tokens.Clause{
			Field:    "age",
			Operator: filter.OpIn,
			Value:    "[20, 30, 40]",
		},
	},
	{
		name:  "in operator with type mismatch in list",
		input: "age in [20, thirty, 40]",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: TypeMismatchError{
			Field:    "age",
			Value:    "[20, thirty, 40]",
			Expected: filter.TypeNumber,
		},
	},
	{
		name:  "in operator missing value",
		input: "name in",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expectedErr: InvalidExpressionError("name in"),
	},
	{
		name:        "field not found - spaces",
		input:       "narm = John",
		fields:      filter.Fields{},
		expectedErr: FieldNotFoundError("narm"),
	},
	{
		name:  "invalid operator - spaces",
		input: "name is Jane",
		fields: filter.Fields{
			"name": filter.TypeString,
		},
		expectedErr: InvalidOperatorError("is"),
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

type testParseGroupTokenTestCase struct {
	name     string
	input    string
	expected tokens.FilterToken
}

var testParseGroupTokenTestCases = []testParseGroupTokenTestCase{
	{
		name:  "simple group token",
		input: "(name='John', name='Jane')",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "Jane",
				},
			},
			JoinOp: tokens.OpAnd,
		},
	},
	{
		name:  "simple OR group token",
		input: "(name='John' OR name='Jane')",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "Jane",
				},
			},
			JoinOp: tokens.OpOr,
		},
	},
	{
		name:  "simple OR token - no parentheses",
		input: "name='John' OR name='Jane'",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "Jane",
				},
			},
			JoinOp: tokens.OpOr,
		},
	},
	{
		name:  "nested group token",
		input: "(name='John' OR (name='Jane' AND city='NY'))",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
			},
			JoinOp: tokens.OpOr,
			Children: []tokens.NextFilterToken{
				{
					Op: tokens.OpOr,
					T: tokens.FilterToken{
						Clauses: []tokens.Clause{
							{
								Field:    "name",
								Operator: filter.OpEq,
								Value:    "Jane",
							},
							{
								Field:    "city",
								Operator: filter.OpEq,
								Value:    "NY",
							},
						},
						JoinOp: tokens.OpAnd,
					},
				},
			},
		},
	},
	{
		name:  "complex nested group token",
		input: "(name='John' OR (name='Jane' AND (city='NY' OR city='LA')))",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
			},
			JoinOp: tokens.OpOr,
			Children: []tokens.NextFilterToken{
				{
					Op: tokens.OpOr,
					T: tokens.FilterToken{
						Clauses: []tokens.Clause{
							{
								Field:    "name",
								Operator: filter.OpEq,
								Value:    "Jane",
							},
						},
						JoinOp: tokens.OpAnd,
						Children: []tokens.NextFilterToken{
							{
								Op: tokens.OpAnd,
								T: tokens.FilterToken{
									Clauses: []tokens.Clause{
										{
											Field:    "city",
											Operator: filter.OpEq,
											Value:    "NY",
										},
										{
											Field:    "city",
											Operator: filter.OpEq,
											Value:    "LA",
										},
									},
									JoinOp: tokens.OpOr,
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name:  "very complex nested group token",
		input: "(name='John' OR (name='Jane' AND (city='NY' OR (city='LA' AND isActive=true))) OR (name='Doe' AND city='Chicago'))",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
			},
			JoinOp: tokens.OpOr,
			Children: []tokens.NextFilterToken{
				{
					Op: tokens.OpOr,
					T: tokens.FilterToken{
						Clauses: []tokens.Clause{
							{
								Field:    "name",
								Operator: filter.OpEq,
								Value:    "Jane",
							},
						},
						JoinOp: tokens.OpAnd,
						Children: []tokens.NextFilterToken{
							{
								Op: tokens.OpAnd,
								T: tokens.FilterToken{
									Clauses: []tokens.Clause{
										{
											Field:    "city",
											Operator: filter.OpEq,
											Value:    "NY",
										},
									},
									JoinOp: tokens.OpOr,
									Children: []tokens.NextFilterToken{
										{
											Op: tokens.OpOr,
											T: tokens.FilterToken{
												Clauses: []tokens.Clause{
													{
														Field:    "city",
														Operator: filter.OpEq,
														Value:    "LA",
													},
													{
														Field:    "isactive",
														Operator: filter.OpEq,
														Value:    "true",
													},
												},
												JoinOp: tokens.OpAnd,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Op: tokens.OpOr,
					T: tokens.FilterToken{
						Clauses: []tokens.Clause{
							{
								Field:    "name",
								Operator: filter.OpEq,
								Value:    "Doe",
							},
							{
								Field:    "city",
								Operator: filter.OpEq,
								Value:    "Chicago",
							},
						},
						JoinOp: tokens.OpAnd,
					},
				},
			},
		},
	},
	{
		name:  "simple spaces, no parentheses",
		input: "name = 'John' OR name = 'Jane'",
		expected: tokens.FilterToken{
			Clauses: []tokens.Clause{
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "John",
				},
				{
					Field:    "name",
					Operator: filter.OpEq,
					Value:    "Jane",
				},
			},
			JoinOp: tokens.OpOr,
		},
	},
}

var groupTokenTestFields = filter.Fields{
	"name":     filter.TypeString,
	"city":     filter.TypeString,
	"isactive": filter.TypeBool,
}

func TestParseGroupToken(t *testing.T) {
	for _, tc := range testParseGroupTokenTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseGroupToken(tc.input, groupTokenTestFields)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !filterTokensAreEqual(result, tc.expected) {
				t.Errorf("expected\n%+v\ngot\n%+v", tc.expected, result)
			}
		})
	}
}

type parseSortTokensTestCase struct {
	name        string
	input       string
	fields      filter.Fields
	expected    tokens.SortToken
	expectedErr error
}

var parseSortTokensTestCases = []parseSortTokensTestCase{
	{
		name:  "simple sort token",
		input: "age:asc",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expected: []tokens.SortClause{
			{
				Field: "age",
				Asc:   true,
			},
		},
	},
	{
		name:  "multiple sort clauses",
		input: "age:asc, name:desc",
		fields: filter.Fields{
			"age":  filter.TypeNumber,
			"name": filter.TypeString,
		},
		expected: []tokens.SortClause{
			{
				Field: "age",
				Asc:   true,
			},
			{
				Field: "name",
				Asc:   false,
			},
		},
	},
	{
		name:  "sort clause with spaces",
		input: "age : asc , name : desc",
		fields: filter.Fields{
			"age":  filter.TypeNumber,
			"name": filter.TypeString,
		},
		expected: []tokens.SortClause{
			{
				Field: "age",
				Asc:   true,
			},
			{
				Field: "name",
				Asc:   false,
			},
		},
	},
	{
		name:        "invalid sort expression - missing colon",
		input:       "age asc",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expectedErr: InvalidSortExpressionError("age asc"),
	},
	{
		name:        "invalid sort expression - too many colons",
		input:       "age:asc:extra",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expectedErr: InvalidSortExpressionError("age:asc:extra"),
	},
	{
		name:        "field not found",
		input:       "height:asc",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expectedErr: FieldNotFoundError("height"),
	},
	{
		name:        "invalid sort direction",
		input:       "age:up",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expectedErr: InvalidSortDirectionError("up"),
	},
	{
		name:        "empty sort expression",
		input:       "   ",
		fields:      filter.Fields{"age": filter.TypeNumber},
		expected:    nil,
		expectedErr: nil,
	},
	{
		name:  "invalid sort expression - missing field",
		input: ":asc",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: InvalidSortExpressionError(":asc"),
	},
	{
		name:  "invalid sort expression - missing direction",
		input: "age:",
		fields: filter.Fields{
			"age": filter.TypeNumber,
		},
		expectedErr: InvalidSortExpressionError("age:"),
	},
}

func TestParseSortTokens(t *testing.T) {
	for _, tc := range parseSortTokensTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseSortTokens(tc.input, tc.fields)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error '%v', got '%v'", tc.expectedErr, err)
			}
			if !sortTokensAreEqual(result, tc.expected) {
				t.Errorf("expected token '%v', got '%v'", tc.expected, result)
			}
		})
	}
}

type parsePaginationTokensTestCase struct {
	name        string
	pagination  string
	perPage     string
	page        string
	cursor      string
	expected    tokens.PaginationToken
	expectedErr []error
}

var parsePaginationTokensTestCases = []parsePaginationTokensTestCase{
	{
		name:       "no pagination",
		pagination: "",
		expected:   tokens.PaginationToken{},
	},
	{
		name:       "valid pagination with page and per_page",
		pagination: "true",
		page:       "2",
		perPage:    "20",
		expected: tokens.PaginationToken{
			Paginate: true,
			Page:     2,
			PerPage:  20,
		},
	},
	{
		name:       "valid pagination with cursor",
		pagination: "true",
		cursor:     "abc123",
		perPage:    "20",
		expected: tokens.PaginationToken{
			Paginate: true,
			Cursor:   "abc123",
			PerPage:  20,
			// Page is not required when using cursor-based pagination, so it can be left at its zero value
			Page: 0,
		},
	},
	{
		name:        "invalid pagination value",
		pagination:  "yes",
		expectedErr: []error{InvalidPaginationError{"pagination", "yes"}},
	},
	{
		name:        "page value is word - invalid", // should we fix this?
		pagination:  "true",
		page:        "two",
		expectedErr: []error{InvalidPaginationError{"page", "two"}},
	},
	{
		name:       "invalid values given",
		pagination: "cool",
		page:       "second",
		perPage:    "gajillion",
		expectedErr: []error{
			InvalidPaginationError{"pagination", "cool"},
			InvalidPaginationError{"page", "second"},
			InvalidPaginationError{"per_page", "gajillion"},
		},
	},
}

func TestParsePaginationTokens(t *testing.T) {
	for _, tc := range parsePaginationTokensTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parsePaginationTokens(tc.pagination, tc.perPage, tc.page, tc.cursor)
			if len(tc.expectedErr) == 0 && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, expectedErr := range tc.expectedErr {
				if !errors.Is(err, expectedErr) {
					t.Errorf("expected error '%v' to be in '%v'", expectedErr, err)
				}
			}
			if !paginationTokensAreEqual(result, tc.expected) {
				t.Errorf("expected token '%v', got '%v'", tc.expected, result)
			}
		})
	}
}

type parseTokensTestCase struct {
	name        string
	query       string
	fields      filter.Fields
	expected    tokens.Tokens
	expectedErr []error
}

var parseTokenFields = filter.Fields{
	"name":     filter.TypeString,
	"age":      filter.TypeNumber,
	"city":     filter.TypeString,
	"isactive": filter.TypeBool,
}

var parseTokensTestCases = []parseTokensTestCase{
	// {
	// 	name:   "full query with all components",
	// 	query:  "filters=age>30, (name='John' OR name='Jane')&sort=age:asc, name:desc&pagination=true&page=2&per_page=20",
	// 	fields: parseTokenFields,
	// 	expected: tokens.Tokens{
	// 		Filter: []tokens.FilterToken{
	// 			{
	// 				Clauses: []tokens.Clause{
	// 					{
	// 						Field:    "age",
	// 						Operator: filter.OpGt,
	// 						Value:    "30",
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Clauses: []tokens.Clause{
	// 					{
	// 						Field:    "name",
	// 						Operator: filter.OpEq,
	// 						Value:    "John",
	// 					},
	// 					{
	// 						Field:    "name",
	// 						Operator: filter.OpEq,
	// 						Value:    "Jane",
	// 					},
	// 				},
	// 				JoinOp: tokens.OpOr,
	// 			},
	// 		},
	// 		Sort: []tokens.SortClause{
	// 			{
	// 				Field: "age",
	// 				Asc:   true,
	// 			},
	// 			{
	// 				Field: "name",
	// 				Asc:   false,
	// 			},
	// 		},
	// 		Pagination: tokens.PaginationToken{
	// 			Paginate: true,
	// 			Page:     2,
	// 			PerPage:  20,
	// 		},
	// 	},
	// },
	// {
	// 	name:   "query with in operator",
	// 	query:  "filters=name in [john, barb, keith], age > 30&sort=name:asc",
	// 	fields: parseTokenFields,
	// 	expected: tokens.Tokens{
	// 		Filter: []tokens.FilterToken{
	// 			{
	// 				Clauses: []tokens.Clause{
	// 					{
	// 						Field:    "name",
	// 						Operator: filter.OpIn,
	// 						Value:    "[john, barb, keith]",
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Clauses: []tokens.Clause{
	// 					{
	// 						Field:    "age",
	// 						Operator: filter.OpGt,
	// 						Value:    "30",
	// 					},
	// 				},
	// 			},
	// 		},
	// 		Sort: []tokens.SortClause{
	// 			{
	// 				Field: "name",
	// 				Asc:   true,
	// 			},
	// 		},
	// 	},
	// },
	{
		name:   "everything is invalid",
		query:  "filters=age>>30, narm = 'John' OR name is 'Jane'&sort=age up&pagination=yes&page=second&per_page=many",
		fields: parseTokenFields,
		expectedErr: []error{
			TypeMismatchError{
				Field:    "age",
				Value:    ">30",
				Expected: filter.TypeNumber,
			},
			FieldNotFoundError("narm"),
			InvalidOperatorError("is"),
			InvalidSortExpressionError("age up"),
			InvalidPaginationError{"pagination", "yes"},
			InvalidPaginationError{"per_page", "many"},
			InvalidPaginationError{"page", "second"},
		},
	},
}

func TestParse(t *testing.T) {
	for _, tc := range parseTokensTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse(tc.query, tc.fields)
			if len(tc.expectedErr) == 0 && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tc.expectedErr) > 0 && err == nil {
				t.Fatalf("expected error, got nil")
			}
			for _, expectedErr := range tc.expectedErr {
				if !errors.Is(err, expectedErr) {
					t.Errorf("expected error '%v' to be in '%v'", expectedErr, err)
				}
			}
			if len(tc.expectedErr) == 0 && !tokensAreEqual(result, tc.expected) {
				t.Errorf("expected\n%+v\ngot\n%+v", tc.expected, result)
			}
		})
	}
}

func clausesAreEqual(a, b tokens.Clause) bool {
	return a.Field == b.Field && a.Operator == b.Operator && a.Value == b.Value
}

func filterTokensAreEqual(a, b tokens.FilterToken) bool {
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
		if a.Children[i].Op != b.Children[i].Op || !filterTokensAreEqual(a.Children[i].T, b.Children[i].T) {
			return false
		}
	}
	return true
}

func sortTokensAreEqual(a, b tokens.SortToken) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func paginationTokensAreEqual(a, b tokens.PaginationToken) bool {
	return a.Paginate == b.Paginate && a.Page == b.Page && a.PerPage == b.PerPage && a.Cursor == b.Cursor
}

func tokensAreEqual(a, b tokens.Tokens) bool {

	if len(a.Filter) != len(b.Filter) {
		return false
	}

	for i := range a.Filter {
		if !filterTokensAreEqual(a.Filter[i], b.Filter[i]) {
			return false
		}
	}

	if !sortTokensAreEqual(a.Sort, b.Sort) {
		return false
	}

	if !paginationTokensAreEqual(a.Pagination, b.Pagination) {
		return false
	}

	return true
}
