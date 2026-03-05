package filter

type FieldType uint8

// we'll follow JSON-like type conventions for now
const (
	TypeNumber FieldType = iota
	TypeString
	TypeBool
	TypeBytes
	TypeUnknown
)

// Fields maps a filter's name to its data type
type Fields map[string]FieldType

func (f Filter) AddField(name string, fieldType FieldType) {
	if f.Fields == nil {
		f.Fields = make(Fields)
	}
	f.Fields[name] = fieldType
}
