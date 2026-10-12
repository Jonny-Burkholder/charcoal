package sql

import (
	"charcoal/charcoal"
	"charcoal/example/full/models"
	"database/sql"
)

type bookRepo struct {
	dsn string
}

func NewBookRepo(dsn string) bookRepo {
	return bookRepo{dsn: dsn}
}

func (r bookRepo) GetBooks(query string) ([]models.Book, error) {
	// pass an empty book to charcoal to automatically parse
	// filterable fields
	filter, err := charcoal.Filter(models.Book{})
	if err != nil {
		return nil, err
	}

	// parse the query string into a filter struct
	statement, err := filter.ToSql(query)
	if err != nil {
		return nil, err
	}

	// execute the SQL query against the database and return the results
	db, err := sql.Open("mysql", r.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// the query is now fully formed - pass it to the database and return the results
	results, err := db.Query(statement)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	books := make([]models.Book, 0)
	for results.Next() {
		var b models.Book
		if err := results.Scan(&b.ID, &b.Title, &b.Author); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
