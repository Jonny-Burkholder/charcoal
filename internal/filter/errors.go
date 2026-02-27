package filter

import "errors"

var (
	ErrInvalidDataType   = errors.New("invalid data type. review the documentation to see supported type")
	ErrMultipleIndirects = errors.New("multiple indirects not supported")
	ErrMaxIndirects      = errors.New("max indirects exceeded")
	ErrNilData           = errors.New("data cannot be nil")
)
