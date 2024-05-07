package mongo

import (
	"fmt"
	"forum/internal/models"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const collectionName = "posts"

func TestCreate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Create", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := storage.Create(post)
		if err != nil {
			t.Error(err)
		}
	})

	mt.Run("AddComment", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := storage.Create(post)
		if err != nil {
			t.Error(err)
		}

		comment := models.Comment{
			ID:      primitive.NewObjectID(),
			Author:  post.Author,
			Body:    "test comment",
			Created: time.Now().In(time.UTC).Round(time.Millisecond),
		}
		post.Comments = append(post.Comments, comment)
		postBson, err := postToBSON(post)
		if err != nil {
			t.Fatal(err)
		}
		response := []primitive.E{
			{
				Key:   "ok",
				Value: 1,
			},
			{
				Key:   "value",
				Value: postBson,
			},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(response...))

		postResponse, err := storage.AddComment(post.ID, &comment)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(postResponse, post) {
			t.Errorf("\nwant: %v\nhave: %v", post, postResponse)
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(primitive.E{Key: "ok", Value: 0}))
		_, err = storage.AddComment(post.ID, &comment)
		if err == nil {
			t.Error("expected error, but was nil")
		}
	})
}

func TestFind(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Find", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()
		postBson, err := postToBSON(post)
		if err != nil {
			t.Fatal(err)
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err = storage.Create(post)
		if err != nil {
			t.Error(err)
		}

		post2 := newTestPost()
		post2.ID = primitive.NewObjectID()
		postBson2, err := postToBSON(post2)
		if err != nil {
			t.Fatal(err)
		}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, postBson)
		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, postBson2)
		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		filter := bson.M{"category": "music"}
		postsFind, err := storage.Find(filter)
		if err != nil {
			t.Error(err)
		}
		expected := []*models.Post{post, post2}
		for i, postFind := range postsFind {
			if !reflect.DeepEqual(postsFind, expected) {
				t.Errorf("\nwant: %v\nhave: %v", expected[i], postFind)
			}
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))
		_, err = storage.Find(filter)
		if err == nil {
			t.Error("expected error, but was nil")
		}
	})

	mt.Run("FindOne", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()
		postBson, err := postToBSON(post)
		if err != nil {
			t.Fatal(err)
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err = storage.Create(post)
		if err != nil {
			t.Error(err)
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, postBson))
		postFind, err := storage.FindOne(post.ID)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(postFind, post) {
			t.Errorf("\nwant: %v\nhave: %v", post, postFind)
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(primitive.E{Key: "ok", Value: 0}))
		_, err = storage.FindOne(post.ID)
		if err == nil {
			t.Error("expected error, but was nil")
		}
	})
}

func TestUpdate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("UpdateOne", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := storage.Create(post)
		if err != nil {
			t.Error(err)
		}

		post.Text = "update text"
		postBson, err := postToBSON(post)
		if err != nil {
			t.Fatal(err)
		}
		response := []primitive.E{
			{
				Key:   "ok",
				Value: 1,
			},
			{
				Key:   "value",
				Value: postBson,
			},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(response...))

		update := bson.M{"$set": bson.M{"text": "update text"}}
		postResponse, err := storage.UpdateOne(post.ID, update)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(postResponse, post) {
			t.Errorf("\nwant: %v\nhave: %v", post, postResponse)
		}
	})
}

func TestDelete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("DeleteComment", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()
		comment := models.Comment{
			ID:      primitive.NewObjectID(),
			Author:  post.Author,
			Body:    "test comment",
			Created: time.Now().In(time.UTC).Round(time.Millisecond),
		}
		post.Comments = append(post.Comments, comment)

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := storage.Create(post)
		if err != nil {
			t.Error(err)
		}

		post.Comments = make([]models.Comment, 0)
		postBson, err := postToBSON(post)
		if err != nil {
			t.Fatal(err)
		}
		response := []primitive.E{
			{
				Key:   "ok",
				Value: 1,
			},
			{
				Key:   "value",
				Value: postBson,
			},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(response...))

		postResponse, err := storage.DeleteComment(post.ID, comment.ID)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(postResponse, post) {
			t.Errorf("\nwant: %v\nhave: %v", post, postResponse)
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(primitive.E{Key: "ok", Value: 0}))
		_, err = storage.DeleteComment(post.ID, comment.ID)
		if err == nil {
			t.Error("expected error, but was nil")
		}
	})

	mt.Run("Delete", func(mt *mtest.T) {
		storage := NewPostStorage(mt.DB, collectionName)

		post := newTestPost()

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := storage.Create(post)
		if err != nil {
			t.Error(err)
		}

		response := []primitive.E{
			{
				Key:   "ok",
				Value: 1,
			},
			{
				Key:   "acknowledged",
				Value: true,
			},
			{
				Key:   "n",
				Value: 1,
			},
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(response...))
		err = storage.Delete(post.ID)
		if err != nil {
			t.Error(err)
		}

		response = []primitive.E{
			{
				Key:   "ok",
				Value: 1,
			},
			{
				Key:   "acknowledged",
				Value: true,
			},
			{
				Key:   "n",
				Value: 0,
			},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(response...))
		err = storage.Delete(post.ID)
		if err != errNoPost {
			t.Error(err)
		}
	})
}

func newTestPost() *models.Post {
	author := models.Author{
		ID:       primitive.NewObjectID(),
		Username: "usertest",
	}
	vote := models.Vote{
		Vote:   1,
		UserID: author.ID,
	}
	post := &models.Post{
		Title:            "test title",
		Text:             "test text",
		Type:             "text",
		Category:         "music",
		Author:           author,
		ID:               primitive.NewObjectID(),
		Created:          time.Now().In(time.UTC).Round(time.Millisecond),
		Comments:         make([]models.Comment, 0),
		Votes:            []models.Vote{vote},
		UpvotePercentage: 100,
		Score:            1,
		Views:            0,
	}
	return post
}

func postToBSON(post *models.Post) (bson.D, error) {
	postBson := bson.D{}
	data, err := bson.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("cant marshal post")
	}
	err = bson.Unmarshal(data, &postBson)
	if err != nil {
		return nil, fmt.Errorf("cant unmarshal post")
	}
	return postBson, nil
}
