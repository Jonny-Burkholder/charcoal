package mongo

import (
	"charcoal/charcoal"
	"charcoal/example/full/models"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepo struct {
	dsn string
}

func NewUserRepo(dsn string) userRepo {
	return userRepo{dsn: dsn}
}

func (r userRepo) GetUsers(query string) ([]models.User, error) {
	// pass an empty user to charcoal to automatically parse
	// filterable fields
	filter, err := charcoal.Filter(models.User{})
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

	collection := db.Database("mydb").Collection("users")
	cursor, err := collection.Find(context.Background(), statement)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	users := make([]models.User, 0)
	for cursor.Next(context.Background()) {
		var u models.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
