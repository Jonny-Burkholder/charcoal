package tokens

// SortToken handles how data is ordered when returned.
// It handles multi-sort
type SortToken []SortClause

type SortClause struct {
	Field string
	Asc   bool
}
