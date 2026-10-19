/*
Package sql is a Charcoal adapter that converts db-agnostic tokens
into parameterized sql query fragments.

The adapter takes a Config (table name, column mappings, join definitions)
and a tokens.Tokens tree, and produces a Result containing separate SQL
fragments for WHERE, ORDER BY, LIMIT/OFFSET, and JOINs, along with
positional bind arguments. All values are parameterized as ? placeholders
— nothing is ever interpolated into the SQL string.

Typical usage:

	cfg := sql.Config{
		Table: "users",
		ColumnMap: map[string]string{
			"order_total": "orders.total",
		},
		Joins: []sql.Join{
			{
				Type:    sql.LeftJoin,
				Table:   "orders",
				OnLeft:  "users.id",
				OnRight: "orders.user_id",
			},
		},
	}

	result, err := sql.Build(cfg, toks)
	// result.Joins  => "LEFT JOIN orders ON users.id = orders.user_id"
	// result.Where  => "WHERE users.email = ? AND orders.total > ?"
	// result.Args   => ["foo@bar.com", 100]
*/
package sql
