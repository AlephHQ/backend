package mongo

import (
	"context"
	"log"
	"ncp/backend/env"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
	mgo, err := mongo.Connect(context.Background(), options.Client().ApplyURI(env.MongoURI()))
	if err != nil {
		log.Panic(err)
	}

	client = mgo
}
