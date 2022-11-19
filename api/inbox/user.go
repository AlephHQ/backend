package inbox

import "go.mongodb.org/mongo-driver/bson/primitive"

type user struct {
	ID               primitive.ObjectID `bson:"_id"`
	Username         string             `bson:"username"`
	InternalPassword string             `bson:"internal_password"`
}
