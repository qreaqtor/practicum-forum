package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	Author           Author             `json:"author" bson:"author"`
	Title            string             `json:"title" bson:"title"`
	Text             string             `json:"text" bson:"text,omitempty" binding:"-"`
	URL              string             `json:"url" bson:"url,omitempty" binding:"-"`
	Type             string             `json:"type" bson:"type"`
	Category         string             `json:"category" bson:"category"`
	Created          time.Time          `json:"created" bson:"created"`
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	Votes            []Vote             `json:"votes" bson:"votes"`
	Comments         []Comment          `json:"comments" bson:"comments"`
}

type PostInput struct {
	Title    string `json:"title" valid:"minstringlength(1)"`
	Text     string `json:"text" valid:"minstringlength(4),optional"`
	URL      string `json:"url" valid:"url,optional"`
	Type     string `json:"type" valid:"in(link|text)"`
	Category string `json:"category" valid:"in(music|funny|videos|programming|news|fashion)"`
}
