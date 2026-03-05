package handlers

import "charcoal/example/full/models"

type UserRepo interface {
	GetUsers(query string) ([]models.User, error)
}

type BookRepo interface {
	GetBooks(query string) ([]models.Book, error)
}
