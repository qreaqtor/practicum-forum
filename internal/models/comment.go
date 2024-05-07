package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID      primitive.ObjectID `json:"id" bson:"id"`
	Author  Author             `json:"author" bson:"author"`
	Body    string             `json:"body" bson:"body"`
	Created time.Time          `json:"created" bson:"created"`
}

type CommentInput struct {
	Body string `json:"comment" valid:"minstringlength(1)"`
}
