package filter

// operator constants
const (
	OpEq = iota
	OpNe
	OpGt
	OpGte
	OpLt
	OpLte
	OpIn
	OpNotIn
	OpLike
	OpNotLike
	OpIsNull
	OpNotNull
)

// operatorMap normalizes both symbolic and keyword operators to canonical keyword form.
var OperatorMap = map[string]int{
	"=":        OpEq,
	"eq":       OpEq,
	"!=":       OpNe,
	"ne":       OpNe,
	">":        OpGt,
	"gt":       OpGt,
	">=":       OpGte,
	"gte":      OpGte,
	"<":        OpLt,
	"lt":       OpLt,
	"<=":       OpLte,
	"lte":      OpLte,
	"in":       OpIn,
	"not_in":   OpNotIn,
	"nin":      OpNotIn,
	"not in":   OpNotIn,
	"like":     OpLike,
	"~":        OpLike,
	"!~":       OpNotLike,
	"not like": OpNotLike,
	"!like":    OpNotLike,
	"nlike":    OpNotLike,
	"is_null":  OpIsNull,
	"null":     OpIsNull,
	"is null":  OpIsNull,
	"not_null": OpNotNull,
	"not null": OpNotNull,
	"!null":    OpNotNull,
	"nnull":    OpNotNull,
}
