package handlers

import "charcoal/example/full/models"

type BookListResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Data    []models.Book `json:"data,omitempty"`
	Total   int           `json:"total,omitempty"`
	Page    int           `json:"page,omitempty"`
	Limit   int           `json:"limit,omitempty"`
}

type UserListResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Data    []models.User `json:"data,omitempty"`
	Total   int           `json:"total,omitempty"`
	Page    int           `json:"page,omitempty"`
	Limit   int           `json:"limit,omitempty"`
}
