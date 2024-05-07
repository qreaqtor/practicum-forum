package mongo

import (
	"context"
	"errors"
	"forum/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	errNoPost = errors.New("no post found")
)

type postStorage struct {
	posts *mongo.Collection
}

func NewPostStorage(db *mongo.Database, collectionName string) *postStorage {
	collection := db.Collection(collectionName)

	return &postStorage{
		posts: collection,
	}
}

// Обновляет пост по  его postID
func (p *postStorage) UpdateOne(postID primitive.ObjectID, update bson.M) (*models.Post, error) {
	ctx := context.Background()
	filter := bson.M{"_id": postID}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	post := &models.Post{}
	err := p.posts.FindOneAndUpdate(ctx, filter, update, options).Decode(post)
	return post, err
}

// Создает новый пост
func (p *postStorage) Create(post *models.Post) error {
	ctx := context.Background()
	_, err := p.posts.InsertOne(ctx, post)
	return err
}

// Удаляет пост по его ID
func (p *postStorage) Delete(postID primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"_id": postID}
	result, err := p.posts.DeleteOne(ctx, filter)
	if result.DeletedCount == 0 {
		return errNoPost
	}
	return err
}

// Возвращает все посты, удовлетворяющие filter.
func (p *postStorage) Find(filter bson.M) ([]*models.Post, error) {
	ctx := context.Background()
	cursor, err := p.posts.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	posts := make([]*models.Post, 0)
	err = cursor.All(ctx, &posts)
	return posts, err
}

// Возвращает элемент по postID
func (p *postStorage) FindOne(postID primitive.ObjectID) (*models.Post, error) {
	ctx := context.Background()
	filter := bson.M{"_id": postID}
	post := &models.Post{}
	if err := p.posts.FindOne(ctx, filter).Decode(post); err != nil {
		return nil, err
	}
	return post, nil
}

// Добавляет комментарий comment к посту с postID
func (p *postStorage) AddComment(postID primitive.ObjectID, comment *models.Comment) (*models.Post, error) {
	ctx := context.Background()
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$push": bson.M{
			"comments": comment,
		},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	post := &models.Post{}
	err := p.posts.FindOneAndUpdate(ctx, filter, update, options).Decode(post)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// Удаляет комментарий comment к посту с postID
func (p *postStorage) DeleteComment(postID, commentID primitive.ObjectID) (*models.Post, error) {
	ctx := context.Background()
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$pull": bson.M{
			"comments": bson.M{"id": commentID},
		},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	post := &models.Post{}
	err := p.posts.FindOneAndUpdate(ctx, filter, update, options).Decode(post)
	if err != nil {
		return nil, err
	}
	return post, nil
}
