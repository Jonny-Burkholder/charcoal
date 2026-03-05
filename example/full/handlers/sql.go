package handlers

import "net/http"

type SQLHandler struct {
	userRepo UserRepo
	bookRepo BookRepo
}

func (h SQLHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery

	users, err := h.userRepo.GetUsers(query)
	if err != nil {
		http.Error(w, "Error fetching users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// encode users to JSON and write to response
}

func (h SQLHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery

	books, err := h.bookRepo.GetBooks(query)
	if err != nil {
		http.Error(w, "Error fetching books: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// encode books to JSON and write to response
}
