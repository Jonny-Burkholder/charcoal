package handlers

import "net/http"

type GraphQLHandler struct {
	userRepo UserRepo
	bookRepo BookRepo
}

func (h GraphQLHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: fill out
}

func (h GraphQLHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	// TODO: fill out
}
