package main

import (
	"charcoal/example/full/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// for "simplicity", we're just going to use two models in this example, but
	// we'll use multiple database backend mocks to show how the filtering works
	// with each of them. They will be separated by path, rather than the normal
	// dependency injection you would see here

	graphQLHandler, mongoHandler, sqlHandler := getHandlers()

	// SQL
	mux.HandleFunc("GET mysql/users", sqlHandler.GetUsers)
	mux.HandleFunc("GET mysql/books", sqlHandler.GetBooks)

	// Mongo
	mux.HandleFunc("GET mongo/users", mongoHandler.GetUsers)
	mux.HandleFunc("GET mongo/books", mongoHandler.GetBooks)

	// GraphQL
	mux.HandleFunc("GET graphql/users", graphQLHandler.GetUsers)
	mux.HandleFunc("GET graphql/books", graphQLHandler.GetBooks)

	http.ListenAndServe(":8080", mux)
}

func getHandlers() (handlers.GraphQLHandler, handlers.MongoHandler, handlers.SQLHandler) {
	return handlers.GraphQLHandler{}, handlers.MongoHandler{}, handlers.SQLHandler{}
}
