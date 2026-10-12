package main

import (
	"charcoal/example/full/handlers"
	"charcoal/example/full/repositories/mongo"
	"charcoal/example/full/repositories/sql"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// for "simplicity", we're just going to use two models in this example, but
	// we'll use multiple database backend mocks to show how the filtering works
	// with each of them. They will be separated by path, rather than the normal
	// dependency injection you would see here

	sqlHandler, mongoHandler := getHandlers()

	// SQL
	mux.HandleFunc("GET mysql/users", sqlHandler.GetUsers)
	mux.HandleFunc("GET mysql/books", sqlHandler.GetBooks)

	// Mongo
	mux.HandleFunc("GET mongo/users", mongoHandler.GetUsers)
	mux.HandleFunc("GET mongo/books", mongoHandler.GetBooks)

	http.ListenAndServe(":8080", mux)
}

func getHandlers() (handlers.SQLHandler, handlers.MongoHandler) {
	mysqlDSN := "user:password@tcp(localhost:3306)/mydb"
	mongoDSN := "mongodb://localhost:27017/mydb"

	bookRepos := getBookRepos(mysqlDSN, mongoDSN)
	userRepos := getUserRepos(mysqlDSN, mongoDSN)

	sqlHandler := handlers.SQLHandler{
		BookRepo: bookRepos["sql"],
		UserRepo: userRepos["sql"],
	}

	mongoHandler := handlers.MongoHandler{
		BookRepo: bookRepos["mongo"],
		UserRepo: userRepos["mongo"],
	}

	return sqlHandler, mongoHandler

}

// could clean this up
func getBookRepos(sqldsn, mongodsn string) map[string]handlers.BookRepo {
	return map[string]handlers.BookRepo{
		"sql":   sql.NewBookRepo(sqldsn),
		"mongo": mongo.NewBookRepo(mongodsn),
	}
}

func getUserRepos(sqldsn, mongodsn string) map[string]handlers.UserRepo {
	return map[string]handlers.UserRepo{
		"sql":   sql.NewUserRepo(sqldsn),
		"mongo": mongo.NewUserRepo(mongodsn),
	}
}
