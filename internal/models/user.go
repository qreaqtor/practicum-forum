package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       string `json:"id" valid:"-"`
	Password string `json:"password" valid:"minstringlength(8)"`
	Username string `json:"username" valid:"-"`
}

type Author struct {
	ID       primitive.ObjectID `json:"id" bson:"id"`
	Username string             `json:"username" bson:"username"`
}
