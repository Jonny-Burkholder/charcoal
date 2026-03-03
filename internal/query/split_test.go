package query

import (
	"charcoal/internal/tokens"
	"reflect"
	"testing"
)

type splitFilterExpressionTestCase struct {
	name     string
	input    string
	expected []string
}

var splitFilterExpressionTestCases = []splitFilterExpressionTestCase{
	{
		name:  "simple epxression",
		input: "age > 30",
		expected: []string{
			"age",
			">",
			"30",
		},
	},
	{
		name:  "expression with parentheses and quotes",
		input: "description = 'This should \"not be split\" (at all)'",
		expected: []string{
			"description",
			"=",
			"'This should \"not be split\" (at all)'",
		},
	},
	{
		name:  "expression with brackets",
		input: "name in [john, barb, keith]",
		expected: []string{
			"name",
			"in",
			"[john, barb, keith]",
		},
	},
	{
		name:  "no whitespace",
		input: "age>30",
		expected: []string{
			"age>30",
		},
	},
}

func TestSplitFilterExpression(t *testing.T) {
	for _, tc := range splitFilterExpressionTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitFilterExpression(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

type splitTopLevelTestCase struct {
	name       string
	input      string
	expected   []string
	expectedOp tokens.JoinOp
}

var splitTopLevelTestCases = []splitTopLevelTestCase{
	{
		name:  "simple group",
		input: "(  age > 30 OR name = 'John')",
		expected: []string{
			"age > 30",
			"name = 'John'",
		},
		expectedOp: tokens.OpOr,
	},
	{
		name:  "simple comma separated group",
		input: "(age > 30, name = 'John ' )",
		expected: []string{
			"age > 30",
			"name = 'John '",
		},
		expectedOp: tokens.OpAnd,
	},
	{
		name:  "simple group, no parentheses",
		input: "age > 30 OR name = 'John'  ",
		expected: []string{
			"age > 30",
			"name = 'John'",
		},
		expectedOp: tokens.OpOr,
	},
	{
		name:  "nested group",
		input: "(age > 30 OR (name = 'John' AND city = 'NY'))",
		expected: []string{
			"age > 30",
			"(name = 'John' AND city = 'NY')",
		},
		expectedOp: tokens.OpOr,
	},
}

func TestSplitTopLevel(t *testing.T) {
	for _, tc := range splitTopLevelTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, op := splitTopLevel(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
			if op != tc.expectedOp {
				t.Errorf("expected operator %v, got %v", tc.expectedOp, op)
			}
		})
	}
}
