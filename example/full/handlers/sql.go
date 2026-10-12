package handlers

import (
	"encoding/json"
	"net/http"
)

type SQLHandler struct {
	UserRepo UserRepo
	BookRepo BookRepo
}

func (h SQLHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery

	users, err := h.UserRepo.GetUsers(query)
	if err != nil {
		http.Error(w, "Error fetching users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// encode users to JSON and write to response
	response := UserListResponse{
		Status: "success",
		Data:   users,
		Total:  len(users),
		Page:   1, // Assuming page 1 for simplicity
		Limit:  len(users),
	}

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func (h SQLHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery

	books, err := h.BookRepo.GetBooks(query)
	if err != nil {
		http.Error(w, "Error fetching books: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// encode books to JSON and write to response
	response := BookListResponse{
		Status: "success",
		Data:   books,
		Total:  len(books),
		Page:   1, // Assuming page 1 for simplicity
		Limit:  len(books),
	}

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
