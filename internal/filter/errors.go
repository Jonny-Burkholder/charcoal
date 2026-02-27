package filter

import "errors"

var (
	ErrMultipleIndirects = errors.New("multiple indirects not supported")
	ErrInvalidDataType   = errors.New("invalid data type. review the documentation to see supported type")
	ErrNilData           = errors.New("data cannot be nil")
)
