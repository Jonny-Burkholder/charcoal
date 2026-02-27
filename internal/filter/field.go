package filter

type fieldType uint8

// we'll follow JSON-like type conventions for now
const (
	TypeNumber fieldType = iota
	TypeString
	TypeBool
	TypeBytes
	TypeUnknown
)

// Fields maps a filter's name to its data type
type Fields map[string]fieldType
