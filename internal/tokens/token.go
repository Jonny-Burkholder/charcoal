package tokens

// FilterToken is a tree struct
type FilterToken struct {
	Field    string
	Operator string // probably a custom operator type in the future
	Value    string
	children map[uint8]FilterToken // uint8 represenst like OR or AND
}
