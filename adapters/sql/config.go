package mysql

// JoinType represents the type of SQL JOIN
type JoinType uint8

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
)

// Join describes a single MySQL JOIN clause. The adapter will only include
// joins whose tables are actually referenced by the token tree.
type Join struct {
	Type    JoinType
	Table   string // target table, e.g. "orders"
	Alias   string // optional alias, e.g. "o" — if empty, Table is used
	OnLeft  string // left side of the ON, e.g. "users.id"
	OnRight string // right side of the ON, e.g. "orders.user_id"
}

// Config holds the "do once" configuration for the MySQL adapter.
// Typically created at startup and reused across requests.
type Config struct {
	// Table is the primary table name, e.g. "users"
	Table string

	// ColumnMap maps filter field names to fully-qualified MySQL column
	// references (e.g. "email" → "users.email"). Fields not present in
	// the map are assumed to match the column name on the primary table,
	// i.e. the filter name "status" maps to "Table.status".
	ColumnMap map[string]string

	// Joins defines the available join clauses. Only joins whose tables
	// are actually referenced by the query's filter/sort tokens will be
	// included in the output. Each join is included at most once.
	Joins []Join
}

// resolveColumn returns the fully-qualified column name for a filter field.
// If the field exists in ColumnMap, that value is returned. Otherwise the
// field name is qualified against the primary table.
func (c Config) resolveColumn(field string) string {
	if col, ok := c.ColumnMap[field]; ok {
		return col
	}
	return c.Table + "." + field
}

// joinTypeSQL returns the SQL keyword for a JoinType
func joinTypeSQL(jt JoinType) string {
	switch jt {
	case LeftJoin:
		return "LEFT JOIN"
	case RightJoin:
		return "RIGHT JOIN"
	default:
		return "JOIN"
	}
}
