package managers

import (
	"errors"
	"forum/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	errBadAction = errors.New("bad action")

	upvoteAction   = "upvote"
	downvoteAction = "downvote"
	unvoteAction   = "unvote"
)

type postRepo interface {
	Find(bson.M) ([]*models.Post, error)
	FindOne(primitive.ObjectID) (*models.Post, error)
	UpdateOne(primitive.ObjectID, bson.M) (*models.Post, error)
	AddComment(primitive.ObjectID, *models.Comment) (*models.Post, error)
	DeleteComment(primitive.ObjectID, primitive.ObjectID) (*models.Post, error)
	Delete(primitive.ObjectID) error
	Create(*models.Post) error
}

type PostManager struct {
	storage postRepo
}

func NewPostManager(storage postRepo) *PostManager {
	return &PostManager{
		storage: storage,
	}
}

// Возвращает все имеющиеся посты по category
func (pm *PostManager) GetAllByCategory(category string) ([]*models.Post, error) {
	filter := bson.M{"category": category}
	return pm.storage.Find(filter)
}

// Возвращает все имеющиеся посты по username
func (pm *PostManager) GetAllByUser(username string) ([]*models.Post, error) {
	filter := bson.M{"author.username": username}
	return pm.storage.Find(filter)
}

// Возвращает все имеющиеся посты
func (pm *PostManager) GetAll() ([]*models.Post, error) {
	filter := bson.M{}
	return pm.storage.Find(filter)
}

// Возвращает пост по postID.
// Используется UpdateOne, а не FindOne, т.к. при запросе поста необходимо увеличивать кол-во просмотров.
func (pm *PostManager) FindOne(postIDStr string) (*models.Post, error) {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return nil, err
	}
	update := bson.M{
		"$inc": bson.M{
			"views": 1,
		},
	}
	return pm.storage.UpdateOne(postID, update)
}

/*
Выполняет действие пользователя action (upvote|downvote|unvote) и перерасчет score, percentUpvote.
*/
func (pm *PostManager) UpdateVotes(postIDStr, action string, authorID primitive.ObjectID) (*models.Post, error) {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return nil, err
	}

	post, err := pm.storage.FindOne(postID)
	if err != nil {
		return nil, err
	}

	votePos := -1
	for i, vote := range post.Votes {
		if vote.UserID == authorID {
			votePos = i
			break
		}
	}

	if votePos == -1 {
		newVote := models.Vote{UserID: authorID}
		post.Votes = append(post.Votes, newVote)
		votePos = len(post.Votes) - 1
	}

	post.Score -= post.Votes[votePos].Vote

	switch action {
	case upvoteAction:
		post.Votes[votePos].Vote = 1
		post.Score += 1
	case downvoteAction:
		post.Votes[votePos].Vote = -1
		post.Score -= 1
	case unvoteAction:
		post.Votes = append(post.Votes[:votePos], post.Votes[votePos+1:]...)
	default:
		return nil, errBadAction
	}

	voteCount := len(post.Votes)
	if voteCount == 0 {
		post.UpvotePercentage = 0
	} else {
		post.UpvotePercentage = (voteCount + post.Score) * 50 / voteCount
	}

	update := bson.M{
		"$set": bson.M{
			"score":            post.Score,
			"upvotePercentage": post.UpvotePercentage,
			"votes":            post.Votes,
		},
	}

	return pm.storage.UpdateOne(postID, update)
}

/*
Создает новый комментарий на основе commentIn к посту с postID.
*/
func (pm *PostManager) AddComment(postIDStr string, commentIn *models.CommentInput, author *models.Author) (*models.Post, error) {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return nil, err
	}

	newComment := &models.Comment{
		ID:      primitive.NewObjectID(),
		Author:  *author,
		Created: time.Now(),
		Body:    commentIn.Body,
	}

	return pm.storage.AddComment(postID, newComment)
}

/*
Удаляет комментарий с id commentID у поста с postID
*/
func (pm *PostManager) DeleteComment(postIDStr, commentIDStr string) (*models.Post, error) {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return nil, err
	}

	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		return nil, err
	}

	return pm.storage.DeleteComment(postID, commentID)
}

// Удаляет пост
func (pm *PostManager) Delete(postIDStr string) error {
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		return err
	}
	return pm.storage.Delete(postID)
}

// Создает пост
func (pm *PostManager) Create(post *models.PostInput, author *models.Author) (*models.Post, error) {
	vote := models.Vote{
		Vote:   1,
		UserID: author.ID,
	}
	newPost := &models.Post{
		Title:            post.Title,
		Text:             post.Text,
		URL:              post.URL,
		Type:             post.Type,
		Category:         post.Category,
		Author:           *author,
		ID:               primitive.NewObjectID(),
		Created:          time.Now(),
		Comments:         make([]models.Comment, 0),
		Votes:            []models.Vote{vote},
		UpvotePercentage: 100,
		Score:            1,
		Views:            0,
	}

	err := pm.storage.Create(newPost)
	if err != nil {
		return nil, err
	}
	return newPost, nil
}
