package sql

import (
	"charcoal/charcoal"
	"charcoal/example/full/models"
	"database/sql"
)

type userRepo struct {
	dsn string
}

func NewUserRepo(dsn string) userRepo {
	return userRepo{dsn: dsn}
}

func (r userRepo) GetUsers(query string) ([]models.User, error) {
	// pass an empty user to charcoal to automatically parse
	// filterable fields
	filter, err := charcoal.Filter(models.User{})
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

	users := make([]models.User, 0)
	for results.Next() {
		var u models.User
		if err := results.Scan(&u.ID, &u.FirstName, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
