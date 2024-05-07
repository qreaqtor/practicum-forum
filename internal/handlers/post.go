package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"forum/internal/handlers/utils"
	"forum/internal/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type postManager interface {
	GetAll() ([]*models.Post, error)
	GetAllByCategory(string) ([]*models.Post, error)
	GetAllByUser(string) ([]*models.Post, error)
	FindOne(postID string) (*models.Post, error)
	Delete(postID string) error
	UpdateVotes(string, string, primitive.ObjectID) (*models.Post, error)
	Create(*models.PostInput, *models.Author) (*models.Post, error)
	DeleteComment(string, string) (*models.Post, error)
	AddComment(string, *models.CommentInput, *models.Author) (*models.Post, error)
}

type PostHandler struct {
	Logger      *slog.Logger
	PostManager postManager
}

// Хендлер, возвращающий все посты
func (ph *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	posts, err := ph.PostManager.GetAll()
	if err != nil {
		msg.Set(err.Error(), http.StatusBadRequest)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, posts)
}

// Хендлер, выполняющий создание поста
func (ph *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	data, err := utils.ReadRequestBody(r)
	if err != nil {
		msg.Set(err.Error(), http.StatusBadRequest)
		utils.WriteError(w, msg)
		return
	}

	post := &models.PostInput{}
	err = json.Unmarshal(data, post)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	err = utils.ValidateStruct(post)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	author, ok := r.Context().Value(models.CtxKey("user")).(*models.Author)
	if !ok {
		msg.Set("bad context value", http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	createdPost, err := ph.PostManager.Create(post, author)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, createdPost)
}

// Хендлер, возвращающиq пост c id - postID
func (ph *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	postID := mux.Vars(r)["postID"]
	post, err := ph.PostManager.FindOne(postID)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, post)
}

// Хендлер, возвращающий все посты по category
func (ph *PostHandler) GetAllByCategory(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	category := mux.Vars(r)["category"]
	posts, err := ph.PostManager.GetAllByCategory(category)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, posts)
}

// Хендлер, возвращающий все посты пользователя по его username
func (ph *PostHandler) GetAllByUser(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	username := mux.Vars(r)["username"]
	posts, err := ph.PostManager.GetAllByUser(username)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, posts)
}

// Хендлер удаления поста с postID
func (ph *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	postID := mux.Vars(r)["postID"]
	err := ph.PostManager.Delete(postID)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, map[string]interface{}{
		"message": "success",
	})
}

// Хендлер, обрабатывающий изменения рейтинга поста c id - postID, action - действие пользователя
func (ph *PostHandler) UpdateVotes(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	vars := mux.Vars(r)
	postID := vars["postID"]
	action := vars["action"]

	author, ok := r.Context().Value(models.CtxKey("user")).(*models.Author)
	if !ok {
		msg.Set("bad context value by key user", http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	post, err := ph.PostManager.UpdateVotes(postID, action, author.ID)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, post)
}

// Хендлер добавления комментариев
func (ph *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	data, err := utils.ReadRequestBody(r)
	if err != nil {
		msg.Set(err.Error(), http.StatusBadRequest)
		utils.WriteError(w, msg)
		return
	}

	comment := &models.CommentInput{}
	err = json.Unmarshal(data, comment)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	err = utils.ValidateStruct(comment)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	postID := mux.Vars(r)["postID"]

	author, ok := r.Context().Value(models.CtxKey("user")).(*models.Author)
	if !ok {
		msg.Set("bad context value by key user", http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	post, err := ph.PostManager.AddComment(postID, comment, author)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, post)
}

// Хендлер удаления комментариев
func (ph *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(ph.Logger, r.URL.Path, r.Method)

	vars := mux.Vars(r)
	postID := vars["postID"]
	commentID := vars["commentID"]
	post, err := ph.PostManager.DeleteComment(postID, commentID)
	if err != nil {
		msg.Set(err.Error(), http.StatusNotFound)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, post)
}
