package main

import "charcoal/charcoal"

type userRepo struct{}

func NewUserRepo() userRepo {
	return userRepo{}
}

func (r userRepo) GetUsers(query string) ([]User, error) {

	filter, err := charcoal.Filter([]User{})
	if err != nil {
		return nil, err
	}

	result := filter.Activate(query)
	if result.Error != nil {
		return nil, result.Error
	}

	return mockDB{}.Query(query)

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
