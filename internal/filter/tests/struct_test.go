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
		input: simple{},
		expected: filter.Fields{
			"Name": filter.TypeString,
			"Age":  filter.TypeNumber,
		},
	},
	{
		name:  "nested struct",
		input: nested{},
		expected: filter.Fields{
			"ID":          filter.TypeString,
			"Simple.Name": filter.TypeString,
			"Simple.Age":  filter.TypeNumber,
		},
	},
	{
		name:  "pointer struct",
		input: pointer{},
		expected: filter.Fields{
			"ID":          filter.TypeString,
			"Simple.Name": filter.TypeString,
			"Simple.Age":  filter.TypeNumber,
		},
		config: &filter.Config{MaxIndirects: 2},
	},
	{
		name:  "multiple indirects struct",
		input: multipleIndirects{},
		expected: filter.Fields{
			"ID":          filter.TypeString,
			"Simple.Name": filter.TypeString,
			"Simple.Age":  filter.TypeNumber,
		},
		config: &filter.Config{MaxIndirects: 3},
	},
	{
		name:   "max indirects exceeded",
		input:  maxIndirects{},
		err:    filter.ErrMaxIndirects,
		config: &filter.Config{MaxIndirects: 2},
	},
	{
		name:  "unexported fields - don't skip",
		input: unexported{},
		expected: filter.Fields{
			"name": filter.TypeString,
			"age":  filter.TypeNumber,
		},
		config: &filter.Config{SkipHiddenFields: false},
	},
	{
		name:     "unexported fields - skip",
		input:    unexported{},
		expected: filter.Fields{},
		config:   &filter.Config{SkipHiddenFields: true},
	},
	{
		name:  "double nested struct",
		input: doubleNested{},
		expected: filter.Fields{
			"ID":                 filter.TypeString,
			"Simple.Name":        filter.TypeString,
			"Simple.Age":         filter.TypeNumber,
			"Nested.ID":          filter.TypeString,
			"Nested.Simple.Name": filter.TypeString,
			"Nested.Simple.Age":  filter.TypeNumber,
		},
	},
}

func TestStruct(t *testing.T) {
	for _, testCase := range structTests {
		t.Run(testCase.name, func(t *testing.T) {
			f, _ := filter.New(testCase.config)
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

			for k, v := range fields {
				t.Logf("field: %s, type: %v", k, v)
			}
		})
	}
}
