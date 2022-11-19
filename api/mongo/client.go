package mongo

import "go.mongodb.org/mongo-driver/mongo"

func Client() *mongo.Client {
	return client
}
