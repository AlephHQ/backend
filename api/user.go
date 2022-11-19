package api

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailAddress struct {
	Addr    string
	Primary bool
}

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	FirstName      string             `bson:"first_name" json:"first_name"`
	LastName       string             `bson:"last_name" json:"last_name"`
	EmailAddresses []EmailAddress     `bson:"email_addresses" json:"email_addresses"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`

	username             string `bson:"username"`
	password             string `bson:"password"`
	internalEmailAddress string `bson:"internal_email_address"`
	internalPassword     string `bson:"internal_password"`
}

func NewUser() *User {
	return &User{}
}

func (u *User) SetPrimaryEmailAddress(email string) *User {
	found := false
	for i, addr := range u.EmailAddresses {
		if addr.Addr == email {
			u.EmailAddresses[i].Primary = true
			found = true
		} else if addr.Primary {
			u.EmailAddresses[i].Primary = false
		}
	}

	if !found {
		u.EmailAddresses = append(u.EmailAddresses, EmailAddress{Addr: email, Primary: true})
	}

	return u
}

func (u *User) SetCreatedAt() *User {
	u.CreatedAt = time.Now()

	return u
}

func (u *User) SetPassword(pw string) *User {
	u.password = pw

	return u
}
