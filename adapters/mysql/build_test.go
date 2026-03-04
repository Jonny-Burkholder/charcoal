package mysql

import (
	"charcoal/internal/filter"
	"charcoal/internal/tokens"
	"errors"
	"testing"
)

type buildTestCase struct {
	name          string
	config        Config
	tokens        tokens.Tokens
	expectedWhere string
	expectedJoins string
	expectedOrder string
	expectedPag   string
	expectedArgs  []any
	expectedErr   error
}

var buildTestCases = []buildTestCase{
	{
		name: "empty tokens produces no output",
		config: Config{
			Table: "users",
		},
		tokens:        tokens.Tokens{},
		expectedWhere: "",
		expectedJoins: "",
		expectedOrder: "",
		expectedPag:   "",
		expectedArgs:  nil,
	},
	{
		name: "single eq clause uses primary table",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "email", Operator: filter.OpEq, Value: "foo@bar.com"},
					},
				},
			},
		},
		expectedWhere: "WHERE users.email = ?",
		expectedArgs:  []any{"foo@bar.com"},
	},
	{
		name: "column map overrides field name",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"email": "u.email_address",
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "email", Operator: filter.OpEq, Value: "foo@bar.com"},
					},
				},
			},
		},
		expectedWhere: "WHERE u.email_address = ?",
		expectedArgs:  []any{"foo@bar.com"},
	},
	{
		name: "multiple clauses joined with AND",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "name", Operator: filter.OpEq, Value: "Alice"},
						{Field: "age", Operator: filter.OpGt, Value: "30"},
					},
					JoinOp: tokens.OpAnd,
				},
			},
		},
		expectedWhere: "WHERE (users.name = ? AND users.age > ?)",
		expectedArgs:  []any{"Alice", "30"},
	},
	{
		name: "multiple clauses joined with OR",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "status", Operator: filter.OpEq, Value: "active"},
						{Field: "status", Operator: filter.OpEq, Value: "pending"},
					},
					JoinOp: tokens.OpOr,
				},
			},
		},
		expectedWhere: "WHERE (users.status = ? OR users.status = ?)",
		expectedArgs:  []any{"active", "pending"},
	},
	{
		name: "IS NULL operator produces no args",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "deleted_at", Operator: filter.OpIsNull},
					},
				},
			},
		},
		expectedWhere: "WHERE users.deleted_at IS NULL",
		expectedArgs:  nil,
	},
	{
		name: "IS NOT NULL operator",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "email", Operator: filter.OpNotNull},
					},
				},
			},
		},
		expectedWhere: "WHERE users.email IS NOT NULL",
		expectedArgs:  nil,
	},
	{
		name: "IN operator splits comma values",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "role", Operator: filter.OpIn, Value: "admin,editor,viewer"},
					},
				},
			},
		},
		expectedWhere: "WHERE users.role IN (?, ?, ?)",
		expectedArgs:  []any{"admin", "editor", "viewer"},
	},
	{
		name: "NOT IN operator",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "role", Operator: filter.OpNotIn, Value: "banned,suspended"},
					},
				},
			},
		},
		expectedWhere: "WHERE users.role NOT IN (?, ?)",
		expectedArgs:  []any{"banned", "suspended"},
	},
	{
		name: "LIKE operator",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "name", Operator: filter.OpLike, Value: "%alice%"},
					},
				},
			},
		},
		expectedWhere: "WHERE users.name LIKE ?",
		expectedArgs:  []any{"%alice%"},
	},
	{
		name: "all comparison operators",
		config: Config{
			Table: "t",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "a", Operator: filter.OpNe, Value: "1"},
					},
				},
				{
					Clauses: []tokens.Clause{
						{Field: "b", Operator: filter.OpGte, Value: "2"},
					},
				},
				{
					Clauses: []tokens.Clause{
						{Field: "c", Operator: filter.OpLt, Value: "3"},
					},
				},
				{
					Clauses: []tokens.Clause{
						{Field: "d", Operator: filter.OpLte, Value: "4"},
					},
				},
			},
		},
		expectedWhere: "WHERE t.a != ? AND t.b >= ? AND t.c < ? AND t.d <= ?",
		expectedArgs:  []any{"1", "2", "3", "4"},
	},
	{
		name: "sort single field ascending",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Sort: tokens.SortToken{
				{Field: "name", Asc: true},
			},
		},
		expectedOrder: "ORDER BY users.name ASC",
	},
	{
		name: "sort multiple fields mixed direction",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Sort: tokens.SortToken{
				{Field: "name", Asc: true},
				{Field: "created_at", Asc: false},
			},
		},
		expectedOrder: "ORDER BY users.name ASC, users.created_at DESC",
	},
	{
		name: "sort uses column map",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"created": "users.created_at",
			},
		},
		tokens: tokens.Tokens{
			Sort: tokens.SortToken{
				{Field: "created", Asc: false},
			},
		},
		expectedOrder: "ORDER BY users.created_at DESC",
	},
	{
		name: "pagination with page and per_page",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Pagination: tokens.PaginationToken{
				Paginate: true,
				Page:     3,
				PerPage:  25,
			},
		},
		expectedPag:  "LIMIT ? OFFSET ?",
		expectedArgs: []any{25, 50},
	},
	{
		name: "pagination page 1 has offset 0",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Pagination: tokens.PaginationToken{
				Paginate: true,
				Page:     1,
				PerPage:  10,
			},
		},
		expectedPag:  "LIMIT ? OFFSET ?",
		expectedArgs: []any{10, 0},
	},
	{
		name: "join auto-included when table referenced in filter",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"order_total": "orders.total",
			},
			Joins: []Join{
				{
					Type:    LeftJoin,
					Table:   "orders",
					OnLeft:  "users.id",
					OnRight: "orders.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "order_total", Operator: filter.OpGt, Value: "100"},
					},
				},
			},
		},
		expectedWhere: "WHERE orders.total > ?",
		expectedJoins: "LEFT JOIN orders ON users.id = orders.user_id",
		expectedArgs:  []any{"100"},
	},
	{
		name: "join auto-pruned when table not referenced",
		config: Config{
			Table: "users",
			Joins: []Join{
				{
					Type:    LeftJoin,
					Table:   "orders",
					OnLeft:  "users.id",
					OnRight: "orders.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "name", Operator: filter.OpEq, Value: "Alice"},
					},
				},
			},
		},
		expectedWhere: "WHERE users.name = ?",
		expectedJoins: "",
		expectedArgs:  []any{"Alice"},
	},
	{
		name: "join included when table referenced in sort",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"order_date": "orders.created_at",
			},
			Joins: []Join{
				{
					Type:    InnerJoin,
					Table:   "orders",
					OnLeft:  "users.id",
					OnRight: "orders.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Sort: tokens.SortToken{
				{Field: "order_date", Asc: false},
			},
		},
		expectedOrder: "ORDER BY orders.created_at DESC",
		expectedJoins: "JOIN orders ON users.id = orders.user_id",
	},
	{
		name: "join with alias",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"order_total": "o.total",
			},
			Joins: []Join{
				{
					Type:    LeftJoin,
					Table:   "orders",
					Alias:   "o",
					OnLeft:  "users.id",
					OnRight: "o.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "order_total", Operator: filter.OpGt, Value: "50"},
					},
				},
			},
		},
		expectedWhere: "WHERE o.total > ?",
		expectedJoins: "LEFT JOIN orders o ON users.id = o.user_id",
		expectedArgs:  []any{"50"},
	},
	{
		name: "duplicate joins only included once",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"order_total":  "orders.total",
				"order_status": "orders.status",
			},
			Joins: []Join{
				{
					Type:    LeftJoin,
					Table:   "orders",
					OnLeft:  "users.id",
					OnRight: "orders.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "order_total", Operator: filter.OpGt, Value: "100"},
						{Field: "order_status", Operator: filter.OpEq, Value: "shipped"},
					},
					JoinOp: tokens.OpAnd,
				},
			},
		},
		expectedWhere: "WHERE (orders.total > ? AND orders.status = ?)",
		expectedJoins: "LEFT JOIN orders ON users.id = orders.user_id",
		expectedArgs:  []any{"100", "shipped"},
	},
	{
		name: "multiple joins only relevant ones included",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"order_total": "orders.total",
			},
			Joins: []Join{
				{
					Type:    LeftJoin,
					Table:   "orders",
					OnLeft:  "users.id",
					OnRight: "orders.user_id",
				},
				{
					Type:    LeftJoin,
					Table:   "profiles",
					OnLeft:  "users.id",
					OnRight: "profiles.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "order_total", Operator: filter.OpGt, Value: "100"},
					},
				},
			},
		},
		expectedWhere: "WHERE orders.total > ?",
		expectedJoins: "LEFT JOIN orders ON users.id = orders.user_id",
		expectedArgs:  []any{"100"},
	},
	{
		name: "full query with filter sort pagination and join",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"order_total": "orders.total",
			},
			Joins: []Join{
				{
					Type:    LeftJoin,
					Table:   "orders",
					OnLeft:  "users.id",
					OnRight: "orders.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "name", Operator: filter.OpLike, Value: "%alice%"},
						{Field: "order_total", Operator: filter.OpGte, Value: "50"},
					},
					JoinOp: tokens.OpAnd,
				},
			},
			Sort: tokens.SortToken{
				{Field: "name", Asc: true},
			},
			Pagination: tokens.PaginationToken{
				Paginate: true,
				Page:     2,
				PerPage:  20,
			},
		},
		expectedWhere: "WHERE (users.name LIKE ? AND orders.total >= ?)",
		expectedJoins: "LEFT JOIN orders ON users.id = orders.user_id",
		expectedOrder: "ORDER BY users.name ASC",
		expectedPag:   "LIMIT ? OFFSET ?",
		expectedArgs:  []any{"%alice%", "50", 20, 20},
	},
	{
		name: "empty table name returns error",
		config: Config{
			Table: "",
		},
		tokens:      tokens.Tokens{},
		expectedErr: ErrEmptyTable,
	},
	{
		name: "unsupported operator returns error",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "x", Operator: 9999, Value: "y"},
					},
				},
			},
		},
		expectedErr: ErrBuild,
	},
	{
		name: "filter token with children",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "active", Operator: filter.OpEq, Value: "true"},
					},
					Children: []tokens.NextFilterToken{
						{
							Op: tokens.OpAnd,
							T: tokens.FilterToken{
								Clauses: []tokens.Clause{
									{Field: "role", Operator: filter.OpEq, Value: "admin"},
									{Field: "role", Operator: filter.OpEq, Value: "editor"},
								},
								JoinOp: tokens.OpOr,
							},
						},
					},
				},
			},
		},
		expectedWhere: "WHERE (users.active = ? AND (users.role = ? OR users.role = ?))",
		expectedArgs:  []any{"true", "admin", "editor"},
	},
	{
		name: "right join",
		config: Config{
			Table: "users",
			ColumnMap: map[string]string{
				"log_action": "audit_log.action",
			},
			Joins: []Join{
				{
					Type:    RightJoin,
					Table:   "audit_log",
					OnLeft:  "users.id",
					OnRight: "audit_log.user_id",
				},
			},
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "log_action", Operator: filter.OpEq, Value: "login"},
					},
				},
			},
		},
		expectedWhere: "WHERE audit_log.action = ?",
		expectedJoins: "RIGHT JOIN audit_log ON users.id = audit_log.user_id",
		expectedArgs:  []any{"login"},
	},
	{
		name: "NOT LIKE operator",
		config: Config{
			Table: "users",
		},
		tokens: tokens.Tokens{
			Filter: []tokens.FilterToken{
				{
					Clauses: []tokens.Clause{
						{Field: "name", Operator: filter.OpNotLike, Value: "%test%"},
					},
				},
			},
		},
		expectedWhere: "WHERE users.name NOT LIKE ?",
		expectedArgs:  []any{"%test%"},
	},
}

func TestBuild(t *testing.T) {
	for _, tc := range buildTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Build(tc.config, tc.tokens)

			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error containing %v, got nil", tc.expectedErr)
				}
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Where != tc.expectedWhere {
				t.Errorf("Where:\n  got:  %q\n  want: %q", result.Where, tc.expectedWhere)
			}

			if result.Joins != tc.expectedJoins {
				t.Errorf("Joins:\n  got:  %q\n  want: %q", result.Joins, tc.expectedJoins)
			}

			if result.OrderBy != tc.expectedOrder {
				t.Errorf("OrderBy:\n  got:  %q\n  want: %q", result.OrderBy, tc.expectedOrder)
			}

			if result.Pagination != tc.expectedPag {
				t.Errorf("Pagination:\n  got:  %q\n  want: %q", result.Pagination, tc.expectedPag)
			}

			if !argsEqual(result.Args, tc.expectedArgs) {
				t.Errorf("Args:\n  got:  %v\n  want: %v", result.Args, tc.expectedArgs)
			}
		})
	}
}

// argsEqual compares two arg slices. Both nil and empty are treated as equal.
func argsEqual(a, b []any) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type resolveColumnTestCase struct {
	name     string
	config   Config
	field    string
	expected string
}

var resolveColumnTestCases = []resolveColumnTestCase{
	{
		name:     "unmapped field uses table prefix",
		config:   Config{Table: "users"},
		field:    "email",
		expected: "users.email",
	},
	{
		name: "mapped field uses configured column",
		config: Config{
			Table:     "users",
			ColumnMap: map[string]string{"email": "u.email_address"},
		},
		field:    "email",
		expected: "u.email_address",
	},
}

func TestResolveColumn(t *testing.T) {
	for _, tc := range resolveColumnTestCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.resolveColumn(tc.field)
			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

type joinTypeSQLTestCase struct {
	name     string
	joinType JoinType
	expected string
}

var joinTypeSQLTestCases = []joinTypeSQLTestCase{
	{name: "inner join", joinType: InnerJoin, expected: "JOIN"},
	{name: "left join", joinType: LeftJoin, expected: "LEFT JOIN"},
	{name: "right join", joinType: RightJoin, expected: "RIGHT JOIN"},
}

func TestJoinTypeSQL(t *testing.T) {
	for _, tc := range joinTypeSQLTestCases {
		t.Run(tc.name, func(t *testing.T) {
			got := joinTypeSQL(tc.joinType)
			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

type inClauseTestCase struct {
	name         string
	col          string
	value        string
	negate       bool
	expectedSQL  string
	expectedArgs []any
}

var inClauseTestCases = []inClauseTestCase{
	{
		name:         "simple IN",
		col:          "users.role",
		value:        "a,b,c",
		negate:       false,
		expectedSQL:  "users.role IN (?, ?, ?)",
		expectedArgs: []any{"a", "b", "c"},
	},
	{
		name:         "NOT IN",
		col:          "users.role",
		value:        "x,y",
		negate:       true,
		expectedSQL:  "users.role NOT IN (?, ?)",
		expectedArgs: []any{"x", "y"},
	},
	{
		name:         "values with spaces are trimmed",
		col:          "t.col",
		value:        " a , b , c ",
		negate:       false,
		expectedSQL:  "t.col IN (?, ?, ?)",
		expectedArgs: []any{"a", "b", "c"},
	},
	{
		name:         "empty value produces false condition",
		col:          "t.col",
		value:        "",
		negate:       false,
		expectedSQL:  "1=0",
		expectedArgs: nil,
	},
	{
		name:         "empty NOT IN produces true condition",
		col:          "t.col",
		value:        "",
		negate:       true,
		expectedSQL:  "1=1",
		expectedArgs: nil,
	},
}

func TestBuildInClause(t *testing.T) {
	for _, tc := range inClauseTestCases {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := buildInClause(tc.col, tc.value, tc.negate)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sql != tc.expectedSQL {
				t.Errorf("SQL:\n  got:  %q\n  want: %q", sql, tc.expectedSQL)
			}
			if !argsEqual(args, tc.expectedArgs) {
				t.Errorf("Args:\n  got:  %v\n  want: %v", args, tc.expectedArgs)
			}
		})
	}
}
