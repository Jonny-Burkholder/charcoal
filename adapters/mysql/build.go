package mysql

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
	"strings"
)

// Result holds the parameterized SQL fragments produced by Build.
// Each fragment is independent so callers can compose their own
// SELECT statement however they like.
type Result struct {
	Joins      string // e.g. "LEFT JOIN orders ON users.id = orders.user_id"
	Where      string // e.g. "WHERE users.email = ? AND orders.total > ?"
	OrderBy    string // e.g. "ORDER BY users.name ASC"
	Pagination string // e.g. "LIMIT ? OFFSET ?"
	Args       []any  // positional bind args across all clauses
}

// operatorSQL maps filter operator constants to their MySQL SQL representation.
// Operators like IN, IS NULL, and IS NOT NULL need special handling and are
// dealt with separately.
var operatorSQL = map[int]string{
	filter.OpEq:      "=",
	filter.OpNe:      "!=",
	filter.OpGt:      ">",
	filter.OpGte:     ">=",
	filter.OpLt:      "<",
	filter.OpLte:     "<=",
	filter.OpLike:    "LIKE",
	filter.OpNotLike: "NOT LIKE",
}

// Build converts a db-agnostic token tree into parameterized MySQL query
// fragments. It returns an error if the config is invalid or a token
// contains an unsupported operator.
func Build(cfg Config, toks tokens.Tokens) (Result, error) {
	if cfg.Table == "" {
		return Result{}, errors.Join(ErrBuild, ErrEmptyTable)
	}

	var result Result
	var buildErr error

	// --- WHERE ---
	if len(toks.Filter) > 0 {
		where, args, err := buildWhere(cfg, toks.Filter)
		if err != nil {
			buildErr = errors.Join(buildErr, err)
		} else if where != "" {
			result.Where = "WHERE " + where
			result.Args = append(result.Args, args...)
		}
	}

	// --- ORDER BY ---
	if len(toks.Sort) > 0 {
		result.OrderBy = buildOrderBy(cfg, toks.Sort)
	}

	// --- PAGINATION ---
	if toks.Pagination.Paginate {
		pag, args := buildPagination(toks.Pagination)
		result.Pagination = pag
		result.Args = append(result.Args, args...)
	}

	// --- JOINS (auto-pruned) ---
	if len(cfg.Joins) > 0 {
		result.Joins = buildJoins(cfg, result)
	}

	if buildErr != nil {
		return Result{}, errors.Join(ErrBuild, buildErr)
	}

	return result, nil
}

// buildWhere walks the filter token tree iteratively and produces a
// WHERE clause string with ? placeholders.
func buildWhere(cfg Config, filterTokens []tokens.FilterToken) (string, []any, error) {
	// top-level filter tokens are joined with AND
	var parts []string
	var allArgs []any
	var whereErr error

	for _, ft := range filterTokens {
		fragment, args, err := buildFilterToken(cfg, ft)
		if err != nil {
			whereErr = errors.Join(whereErr, err)
			continue
		}
		if fragment != "" {
			parts = append(parts, fragment)
			allArgs = append(allArgs, args...)
		}
	}

	if whereErr != nil {
		return "", nil, whereErr
	}

	return strings.Join(parts, " AND "), allArgs, nil
}

// stackEntry is used for iterative tree traversal of FilterToken
type stackEntry struct {
	token   tokens.FilterToken
	joinOp  tokens.JoinOp // how this node connects to its parent
	isChild bool          // whether this was pushed as a child (vs root)
}

// buildFilterToken converts a single FilterToken (and its children)
// into a SQL fragment. Uses an explicit stack instead of recursion.
func buildFilterToken(cfg Config, root tokens.FilterToken) (string, []any, error) {
	// For a leaf node with no children, just build the clauses directly
	if len(root.Children) == 0 {
		return buildClauses(cfg, root.Clauses, root.JoinOp)
	}

	// For nodes with children, we need iterative tree traversal.
	// We use a two-pass approach: first flatten the tree into a
	// postfix-ordered list, then build SQL from that.

	type workItem struct {
		token  tokens.FilterToken
		joinOp tokens.JoinOp // how this item connects to its predecessor
	}

	// Flatten using a stack. We process the tree level by level,
	// collecting all fragments, then join them.
	var fragments []string
	var allArgs []any
	var flatErr error

	// Build this node's own clauses first
	if len(root.Clauses) > 0 {
		frag, args, err := buildClauses(cfg, root.Clauses, root.JoinOp)
		if err != nil {
			return "", nil, err
		}
		if frag != "" {
			fragments = append(fragments, frag)
			allArgs = append(allArgs, args...)
		}
	}

	// Now iterate over children using a stack (no recursion)
	type stackItem struct {
		children []tokens.NextFilterToken
		index    int
	}

	stack := []stackItem{{children: root.Children, index: 0}}

	for len(stack) > 0 {
		top := &stack[len(stack)-1]

		if top.index >= len(top.children) {
			// pop
			stack = stack[:len(stack)-1]
			continue
		}

		child := top.children[top.index]
		top.index++

		// Build this child's clauses
		if len(child.T.Clauses) > 0 {
			frag, args, err := buildClauses(cfg, child.T.Clauses, child.T.JoinOp)
			if err != nil {
				flatErr = errors.Join(flatErr, err)
				continue
			}
			if frag != "" {
				joinWord := joinOpSQL(child.Op)
				fragments = append(fragments, joinWord+" "+frag)
				allArgs = append(allArgs, args...)
			}
		}

		// If this child has its own children, push them onto the stack
		if len(child.T.Children) > 0 {
			stack = append(stack, stackItem{children: child.T.Children, index: 0})
		}
	}

	if flatErr != nil {
		return "", nil, flatErr
	}

	if len(fragments) == 0 {
		return "", nil, nil
	}

	// If there's only one fragment, no need for outer parens
	if len(fragments) == 1 {
		return fragments[0], allArgs, nil
	}

	// Multiple fragments already carry their join words from children
	return "(" + strings.Join(fragments, " ") + ")", allArgs, nil
}

// buildClauses converts a slice of Clause into a SQL fragment joined by the given JoinOp.
// e.g. "users.name = ? AND users.age > ?"
func buildClauses(cfg Config, clauses []tokens.Clause, joinOp tokens.JoinOp) (string, []any, error) {
	if len(clauses) == 0 {
		return "", nil, nil
	}

	var parts []string
	var args []any

	for _, c := range clauses {
		col := cfg.resolveColumn(c.Field)
		frag, clauseArgs, err := buildClause(col, c)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, frag)
		args = append(args, clauseArgs...)
	}

	joiner := joinOpSQL(joinOp)
	result := strings.Join(parts, " "+joiner+" ")

	// wrap in parens if multiple clauses to preserve grouping
	if len(parts) > 1 {
		result = "(" + result + ")"
	}

	return result, args, nil
}

// buildClause converts a single Clause into a SQL fragment + args.
func buildClause(col string, c tokens.Clause) (string, []any, error) {
	switch c.Operator {
	case filter.OpIsNull:
		return col + " IS NULL", nil, nil

	case filter.OpNotNull:
		return col + " IS NOT NULL", nil, nil

	case filter.OpIn:
		return buildInClause(col, c.Value, false)

	case filter.OpNotIn:
		return buildInClause(col, c.Value, true)

	default:
		sql, ok := operatorSQL[c.Operator]
		if !ok {
			return "", nil, UnsupportedOperatorError{Operator: c.Operator}
		}
		return col + " " + sql + " ?", []any{c.Value}, nil
	}
}

// buildInClause splits comma-separated values and produces
// "col IN (?, ?, ?)" or "col NOT IN (?, ?, ?)" with corresponding args.
func buildInClause(col string, value string, negate bool) (string, []any, error) {
	rawValues := strings.Split(value, ",")
	var trimmed []string
	for _, v := range rawValues {
		s := strings.TrimSpace(v)
		if s != "" {
			trimmed = append(trimmed, s)
		}
	}

	if len(trimmed) == 0 {
		// empty IN list — produce a condition that's always false / always true
		if negate {
			return "1=1", nil, nil
		}
		return "1=0", nil, nil
	}

	placeholders := make([]string, len(trimmed))
	args := make([]any, len(trimmed))
	for i, v := range trimmed {
		placeholders[i] = "?"
		args[i] = v
	}

	keyword := "IN"
	if negate {
		keyword = "NOT IN"
	}

	return col + " " + keyword + " (" + strings.Join(placeholders, ", ") + ")", args, nil
}

// buildOrderBy produces an ORDER BY clause from sort tokens.
func buildOrderBy(cfg Config, sort tokens.SortToken) string {
	if len(sort) == 0 {
		return ""
	}

	parts := make([]string, len(sort))
	for i, s := range sort {
		col := cfg.resolveColumn(s.Field)
		dir := "ASC"
		if !s.Asc {
			dir = "DESC"
		}
		parts[i] = col + " " + dir
	}

	return "ORDER BY " + strings.Join(parts, ", ")
}

// buildPagination produces a LIMIT/OFFSET clause from pagination tokens.
func buildPagination(p tokens.PaginationToken) (string, []any) {
	if p.PerPage <= 0 {
		return "", nil
	}

	offset := 0
	if p.Page > 1 {
		offset = (p.Page - 1) * p.PerPage
	}

	return "LIMIT ? OFFSET ?", []any{p.PerPage, offset}
}

// buildJoins auto-prunes the configured joins to include only those
// whose tables are actually referenced in the WHERE, ORDER BY, or
// other output fragments. Each join is included at most once.
func buildJoins(cfg Config, result Result) string {
	// Collect all SQL text that might reference table names
	combined := result.Where + " " + result.OrderBy

	seen := make(map[string]bool)
	var parts []string

	for _, j := range cfg.Joins {
		tableName := j.Table
		if j.Alias != "" {
			tableName = j.Alias
		}

		// skip if already included
		if seen[j.Table] {
			continue
		}

		// check if this join's table is referenced in the output
		// we look for "table." pattern to avoid false substring matches
		searchName := tableName + "."
		if !strings.Contains(combined, searchName) {
			continue
		}

		seen[j.Table] = true

		keyword := joinTypeSQL(j.Type)
		alias := ""
		if j.Alias != "" {
			alias = " " + j.Alias
		}
		parts = append(parts, keyword+" "+j.Table+alias+" ON "+j.OnLeft+" = "+j.OnRight)
	}

	return strings.Join(parts, " ")
}

// joinOpSQL returns the SQL keyword for a JoinOp
func joinOpSQL(op tokens.JoinOp) string {
	if op == tokens.OpOr {
		return "OR"
	}
	return "AND"
}
