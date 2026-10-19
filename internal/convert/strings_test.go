package convert

import (
	"fmt"
	"testing"
)

func TestLowercaseAndUnderscore(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name,
		input,
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "trim everything", // maybe this is a bug
			input:    "_ _",
			expected: "",
		},
		{
			name:     "already snake case",
			input:    "this_field_is_snake",
			expected: "this_field_is_snake",
		},
		{
			name:     "camelCase",
			input:    "camelCase",
			expected: "camel_case",
		},
		{
			name:     "PascalCase",
			input:    "PascalCase",
			expected: "pascal_case",
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			output := lowercaseAndUnderscore(tc.input)
			if output != tc.expected {
				t.Fatalf("wanted %s, got %s", tc.expected, output)
			}
		})
	}
}

// TODO: write these test cases
func TestToSnakeCase(t *testing.T) {}
