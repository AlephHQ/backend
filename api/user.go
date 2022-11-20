package api

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailAddress struct {
	Addr    string `bson:"addr" json:"addr"`
	Primary bool   `bson:"primary" json:"primary"`
}

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	FirstName        string             `bson:"first_name" json:"first_name"`
	LastName         string             `bson:"last_name" json:"last_name"`
	EmailAddresses   []EmailAddress     `bson:"email_addresses" json:"email_addresses"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	Username         string             `bson:"username" json:"username"`
	Password         string             `bson:"password" json:"omitempty"`
	InternalPassword string             `bson:"internal_password" json:"omitempty"`
}

func (u *User) JSON() string {
	marshal := &User{
		ID:             u.ID,
		Name:           u.Name,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		EmailAddresses: u.EmailAddresses,
		CreatedAt:      u.CreatedAt,
		Username:       u.Username,
	}

	b, _ := json.Marshal(marshal)
	return string(b)
}
