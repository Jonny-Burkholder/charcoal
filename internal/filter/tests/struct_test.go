package tests

import (
	"charcoal/internal/filter"
	"testing"
)

type structTest struct {
	name     string
	input    any
	config   *filter.Config
	expected filter.Fields
	err      error
}

var structTests = []structTest{
	{
		name:  "simple struct",
		input: simple{Name: "John", Age: 30},
		expected: filter.Fields{
			"Name": filter.TypeString,
			"Age":  filter.TypeNumber,
		},
	},
	{
		name:  "nested struct",
		input: nested{ID: "123", Simple: simple{Name: "John", Age: 30}},
		expected: filter.Fields{
			"ID":          filter.TypeString,
			"Simple.Name": filter.TypeString,
			"Simple.Age":  filter.TypeNumber,
		},
	},
	{
		name:  "pointer struct",
		input: pointer{ID: "123", Simple: &simple{Name: "John", Age: 30}},
		expected: filter.Fields{
			"ID":          filter.TypeString,
			"Simple.Name": filter.TypeString,
			"Simple.Age":  filter.TypeNumber,
		},
	},
	{
		name:  "unexported fields",
		input: unexported{name: "John", age: 30},
		expected: filter.Fields{
			"name": filter.TypeString,
			"age":  filter.TypeNumber,
		},
		config: &filter.Config{SkipHiddenFields: true},
	},
}

func TestStruct(t *testing.T) {
	for _, testCase := range structTests {
		t.Run(testCase.name, func(t *testing.T) {
			f := filter.New(testCase.config)
			fields, err := f.Struct(testCase.input)
			if err != testCase.err {
				t.Fatalf("expected error %v, got %v", testCase.err, err)
			}

			if len(fields) != len(testCase.expected) {
				t.Errorf("expected %d fields, got %d", len(testCase.expected), len(fields))
			}

			for key, expectedType := range testCase.expected {
				if fields[key] != expectedType {
					t.Errorf("expected field %s to be type %v, got %v", key, expectedType, fields[key])
				}
			}
		})
	}
}
