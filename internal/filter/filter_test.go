package filter

import "testing"

type isByteSliceTest struct {
	name     string
	obj      any
	expected bool
}

var isByteSliceTests = []isByteSliceTest{
	{
		name:     "nil",
		obj:      nil,
		expected: false,
	},
	{
		name:     "not a slice",
		obj:      123,
		expected: false,
	},
	{
		name:     "slice of ints",
		obj:      []int{1, 2, 3},
		expected: false,
	},
	{
		name:     "slice of bytes",
		obj:      []byte{1, 2, 3},
		expected: true,
	},
}

func TestIsByteSlice(t *testing.T) {
	for _, tt := range isByteSliceTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isByteSlice(tt.obj); got != tt.expected {
				t.Errorf("isByteSlice(%v) = %v, want %v", tt.obj, got, tt.expected)
			}
		})
	}
}
