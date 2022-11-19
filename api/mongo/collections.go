package mongo

import "go.mongodb.org/mongo-driver/mongo"

const (
	database = "prod"

	collectionNameAuth = "auth"
)

func AuthCollection() *mongo.Collection {
	return Client().Database(database).Collection(collectionNameAuth)
}
