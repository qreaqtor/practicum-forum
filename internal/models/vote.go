package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Vote struct {
	Vote   int                `json:"vote" bson:"vote"`
	UserID primitive.ObjectID `json:"user" bson:"userID"`
}
