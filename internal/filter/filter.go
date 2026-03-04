package filter

import (
	"charcoal/internal/tokens"
	"encoding/json"
	"reflect"
)

type dataKind uint8

const (
	kindStruct dataKind = iota
	kindJSON
	kindInvalid
)

type Filter struct {
	config Config
	Fields Fields
}

// New returns a new filter object, which
// holds config info. All parsing interactions
// are done through the filter object. I was
// very tempted to call this "Activate", but
// I'll have to find something else for that
func New(data any, config ...Config) (Filter, error) {
	var cfg Config
	if len(config) == 0 {
		cfg = defaultConfig()
	} else {
		cfg = config[0]
		if err := cfg.validate(); err != nil {
			panic(err)
		}
	}
	f := Filter{
		config: cfg,
	}

	kind, err := autoDetect(data)
	if err != nil {
		return Filter{}, err
	}

	switch kind {
	case kindStruct:
		fields, err := f.Struct(data)
		if err != nil {
			return Filter{}, err
		}
		f.Fields = fields
	case kindJSON:
		fields, err := f.JSON(data)
		if err != nil {
			return Filter{}, err
		}
		f.Fields = fields
	default:
		panic("unreachable") // TODO: better error handling here
	}

	return f, nil
}

func (f Filter) Activate(query string) Result {
	// TODO: import cycle. I may have to flatten the structure, which is probably for the best anyway
	return Result{
		Tokens: tokens.Tokens{},
		Error:  nil,
	}
}

func autoDetect(obj any) (dataKind, error) {
	k := reflect.TypeOf(obj).Kind()

	if k == reflect.Pointer || k == reflect.Ptr {
		// if the object is nil, that's not valid

		k = reflect.TypeOf(obj).Elem().Kind()
	}

	switch k {
	case reflect.Struct:
		return kindStruct, nil
	case reflect.Slice:
		if isJSON(obj) {
			return kindJSON, nil
		}
		fallthrough
	case reflect.Pointer:
		return kindInvalid, ErrMultipleIndirects
	default:
		return kindInvalid, ErrInvalidDataType
	}
}

// if an object is a pointer, it's not a byte slice. Period
func isByteSlice(obj any) bool {
	if obj == nil {
		return false
	}

	t := reflect.TypeOf(obj)

	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}

// really this should only be byte slices, but in case
// someone is crazy enough to put a string in there, we'll check for that too
func isJSON(obj any) bool {
	if obj == nil {
		return false
	}

	if isByteSlice(obj) {
		return byteSliceIsJSON(obj)
	}

	if s, ok := obj.(string); ok {
		return json.Valid([]byte(s))
	}

	return false
}

func byteSliceIsJSON(obj any) bool {
	b, ok := obj.([]byte)
	if !ok {
		return false
	}

	return json.Valid(b)
}
