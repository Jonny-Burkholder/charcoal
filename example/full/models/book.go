package models

type Book struct {
	ID          int
	Title       string
	Author      string
	Genre       string
	ReleaseDate int64 // Unix timestamp
	Pages       int
}
