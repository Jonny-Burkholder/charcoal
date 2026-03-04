/*
Package mysql is a Charcoal adapter that converts db-agnostic tokens
into parameterized MySQL query fragments.

The adapter takes a Config (table name, column mappings, join definitions)
and a tokens.Tokens tree, and produces a Result containing separate SQL
fragments for WHERE, ORDER BY, LIMIT/OFFSET, and JOINs, along with
positional bind arguments. All values are parameterized as ? placeholders
— nothing is ever interpolated into the SQL string.

Typical usage:

	cfg := mysql.Config{
		Table: "users",
		ColumnMap: map[string]string{
			"order_total": "orders.total",
		},
		Joins: []mysql.Join{
			{
				Type:    mysql.LeftJoin,
				Table:   "orders",
				OnLeft:  "users.id",
				OnRight: "orders.user_id",
			},
		},
	}

	result, err := mysql.Build(cfg, toks)
	// result.Joins  => "LEFT JOIN orders ON users.id = orders.user_id"
	// result.Where  => "WHERE users.email = ? AND orders.total > ?"
	// result.Args   => ["foo@bar.com", 100]
*/
package mysql
