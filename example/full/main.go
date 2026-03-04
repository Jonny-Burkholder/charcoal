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
	mux.Handle("GET mysql/users", sqlHandler.GetUsers)
	mux.Handle("GET mysql/books", sqlHandler.GetBooks)

	// Mongo
	mux.Handle("GET mongo/users", mongoHandler.GetUsers)
	mux.Handle("GET mongo/books", mongoHandler.GetBooks)

	// GraphQL
	mux.Handle("GET graphql/users", graphQLHandler.ServeHTTP)
	mux.Handle("GET graphql/books", graphQLHandler.ServeHTTP)

	http.ListenAndServe(":8080", mux)
}

func getHandlers() (handlers.GraphQLHandler, handlers.MongoHandler, handlers.SQLHandler) {
	return handlers.GraphQLHandler{}, handlers.MongoHandler{}, handlers.SQLHandler{}
}
