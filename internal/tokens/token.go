package tokens

type JoinOp uint8

const (
	OpAnd JoinOp = iota
	OpOr
)

type Tokens struct {
	Filter     []FilterToken
	Sort       SortToken
	Pagination PaginationToken
}
