package mongo

import (
	"charcoal/charcoal"
	"charcoal/example/full/models"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type bookRepo struct {
	dsn string
}

func NewBookRepo(dsn string) bookRepo {
	return bookRepo{dsn: dsn}
}

func (r bookRepo) GetBooks(query string) ([]models.Book, error) {
	// pass an empty book to charcoal to automatically parse
	// filterable fields
	filter, err := charcoal.Filter(models.Book{})
	if err != nil {
		return nil, err
	}

	// parse the query string into a filter struct
	statement, err := filter.ToMongo(query)
	if err != nil {
		return nil, err
	}

	// execute the MongoDB query against the database and return the results
	db, err := mongo.Connect(context.Background(), options.Client().ApplyURI(r.dsn))
	if err != nil {
		return nil, err
	}
	defer db.Disconnect(context.Background())

	collection := db.Database("mydb").Collection("books")
	cursor, err := collection.Find(context.Background(), statement)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	books := make([]models.Book, 0)
	for cursor.Next(context.Background()) {
		var b models.Book
		if err := cursor.Decode(&b); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
