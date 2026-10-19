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
// TODO: replace with sync.Map
type Fields map[string]FieldType

// AddField adds a filterable field to the charcoal filter. It takes
// a string and field type and returns no error. It is non-idempotent
// and as of this version not safe for concurrent operation
func (f Filter) AddField(name string, fieldType FieldType) {
	if f.Fields == nil {
		f.Fields = make(Fields)
	}
	f.Fields[name] = fieldType
}
