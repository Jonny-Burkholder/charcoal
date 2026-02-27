package filter

import (
	"fmt"
	"reflect"
	"testing"
)

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

type removeIndirectsTest struct {
	name     string
	obj      reflect.Type
	expected reflect.Type
	err      error
}

var removeIndirectsTests = []removeIndirectsTest{
	{
		name:     "not a pointer",
		obj:      reflect.TypeFor[int](),
		expected: reflect.TypeFor[int](),
	},
	{
		name:     "pointer to int",
		obj:      reflect.TypeFor[*int](),
		expected: reflect.TypeFor[int](),
	},
	{
		name:     "pointer to pointer to int",
		obj:      reflect.TypeFor[**int](),
		expected: reflect.TypeFor[int](),
	},
	{
		name: "exceeds max indirects",
		obj:  reflect.TypeFor[***int](),
		err:  ErrMaxIndirects,
	},
}

func TestRemoveIndirects(t *testing.T) {
	f := New(&Config{MaxIndirects: 2})
	fmt.Println("max indirects:", f.config.MaxIndirects)

	for _, tt := range removeIndirectsTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.removeIndirects(tt.obj)
			if err != tt.err {
				t.Errorf("removeIndirects(%v) error = %v, want %v", tt.obj, err, tt.err)
				return
			}
			if got != tt.expected {
				t.Errorf("removeIndirects(%v) = %v, want %v", tt.obj, got, tt.expected)
			}
		})
	}
}
