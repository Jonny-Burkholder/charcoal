package tokens

// FilterToken is a tree struct
type FilterToken struct {
	Clauses  []Clause
	JoinOp   JoinOp // JoinOp is what joins the clauses - they can't be mixed for a single clause
	Children []NextFilterToken
}

type Clause struct {
	Field    string
	Operator int
	Value    string
}

type NextFilterToken struct {
	Op JoinOp
	T  FilterToken
}
