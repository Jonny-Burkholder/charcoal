package main

import (
	"charcoal/charcoal"
	"fmt"
)

type userRepo struct {
	db mockDB
}

func NewUserRepo() userRepo {
	return userRepo{}
}

func (r userRepo) GetUsers(query string) ([]User, error) {

	filter, err := charcoal.Filter([]User{})
	if err != nil {
		return nil, err
	}

	sql, err := filter.ToSql(query)
	if err != nil {
		return nil, fmt.Errorf("error converting query to SQL: %w", err)
	}

	return r.db.Query(sql)

}

type mockDB struct{}

func (db mockDB) Query(query string) ([]User, error) {
	return []User{
		{
			Name:     "Harry",
			Age:      30,
			Password: "password",
			Profile: profile{
				Occupation: "Wizard",
			},
		},
		{
			Name:     "Bob",
			Age:      3000,
			Password: "password",
			Profile: profile{
				Occupation: "Magical Assistant/Air Spirit",
				Nickname:   "Bob the Skull",
			},
		},
	}, nil
}
