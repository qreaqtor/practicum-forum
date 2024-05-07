package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"forum/internal/handlers/utils"
	"forum/internal/models"
	"reflect"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreate(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	path := "/api/posts"
	method := http.MethodPost
	handler := postHandler.Create

	// good response
	author := getDefaultAuthor()
	post := getDefaultPost(author)
	postInput := &models.PostInput{
		Title:    post.Title,
		Text:     post.Text,
		Type:     post.Type,
		Category: post.Category,
	}
	body, err := json.Marshal(post)
	if err != nil {
		t.Fatal(err)
	}
	postManager.EXPECT().Create(postInput, author).Return(post, nil)

	request := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(request.Context(), models.CtxKey("user"), author)

	response := &models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err = utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if !reflect.DeepEqual(response, post) {
		t.Errorf("\nwant: %v\nhave: %v", post, response)
	}

	// Context error
	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Create error
	body, err = json.Marshal(post)
	if err != nil {
		t.Fatal(err)
	}
	postManager.EXPECT().Create(postInput, author).Return(nil, fmt.Errorf("error when trye to create post"))

	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(request.Context(), models.CtxKey("user"), author)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Validate, marshall, read request
	badValidItem := &models.PostInput{
		Type:     "audio",
		Category: "study",
	}
	err = utils.CheckCasesWithoutManager(handler, path, method, badValidItem)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddComment(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)

	path := "/api/post/" + post.ID.Hex()
	method := http.MethodPost
	handler := postHandler.AddComment

	// good response
	commentInput := &models.CommentInput{
		Body: "test text",
	}
	comment := models.Comment{
		ID:      primitive.NewObjectID(),
		Author:  *author,
		Created: time.Now().In(time.UTC).Round(time.Millisecond),
		Body:    commentInput.Body,
	}
	post.Comments = append(post.Comments, comment)
	body, err := json.Marshal(commentInput)
	if err != nil {
		t.Fatal(err)
	}
	postManager.EXPECT().AddComment(post.ID.Hex(), commentInput, author).Return(post, nil)

	request := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	vars := map[string]string{
		"postID": post.ID.Hex(),
	}
	request = mux.SetURLVars(request, vars)

	ctx := context.WithValue(request.Context(), models.CtxKey("user"), author)

	response := &models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err = utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if !reflect.DeepEqual(response, post) {
		t.Errorf("\nwant: %v\nhave: %v", post, response)
	}

	// Context error by key vars
	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Context error by key user
	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request = mux.SetURLVars(request, vars)

	ctx = context.WithValue(request.Context(), models.CtxKey("user"), "bad value")

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Add error
	postManager.EXPECT().AddComment(post.ID.Hex(), commentInput, author).Return(nil, fmt.Errorf("error when try to add comment"))

	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request = mux.SetURLVars(request, vars)

	ctx = context.WithValue(request.Context(), models.CtxKey("user"), author)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Validate, marshall, read request
	badValidItem := &models.PostInput{
		Type:     "audio",
		Category: "study",
	}
	err = utils.CheckCasesWithoutManager(handler, path, method, badValidItem)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateVotes(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)

	action := "unvote"
	path := fmt.Sprintf("/api/post/%s/%s", post.ID.Hex(), action)
	method := http.MethodGet
	handler := postHandler.UpdateVotes

	// good response
	post.UpvotePercentage = 0
	post.Score = 0
	post.Votes = []models.Vote{}

	postManager.EXPECT().UpdateVotes(post.ID.Hex(), action, author.ID).Return(post, nil)

	request := httptest.NewRequest(method, path, nil)
	vars := map[string]string{
		"postID": post.ID.Hex(),
		"action": action,
	}
	request = mux.SetURLVars(request, vars)

	ctx := context.WithValue(request.Context(), models.CtxKey("user"), author)

	response := &models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if !reflect.DeepEqual(response, post) {
		t.Errorf("\nwant: %v\nhave: %v", post, response)
	}

	// Context error by key vars
	request = httptest.NewRequest(method, path, nil)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Context error by key user
	request = httptest.NewRequest(method, path, nil)
	request = mux.SetURLVars(request, vars)

	ctx = context.WithValue(request.Context(), models.CtxKey("user"), "bad value")

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Update votes error
	postManager.EXPECT().UpdateVotes(post.ID.Hex(), action, author.ID).Return(nil, fmt.Errorf("some error when try to update"))

	request = httptest.NewRequest(method, path, nil)
	request.Header.Set("Content-Type", "application/json")
	request = mux.SetURLVars(request, vars)

	ctx = context.WithValue(request.Context(), models.CtxKey("user"), author)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request.WithContext(ctx),
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

func TestDeleteComment(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)
	comment := models.Comment{
		ID:      primitive.NewObjectID(),
		Author:  *author,
		Created: time.Now().In(time.UTC).Round(time.Millisecond),
		Body:    "test comment",
	}

	path := fmt.Sprintf("/api/post/%s/%s", post.ID.Hex(), comment.ID.Hex())
	method := http.MethodDelete
	handler := postHandler.DeleteComment

	// good response
	postManager.EXPECT().DeleteComment(post.ID.Hex(), comment.ID.Hex()).Return(post, nil)

	request := httptest.NewRequest(method, path, nil)

	vars := map[string]string{
		"postID":    post.ID.Hex(),
		"commentID": comment.ID.Hex(),
	}
	request = mux.SetURLVars(request, vars)

	response := &models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if !reflect.DeepEqual(response, post) {
		t.Errorf("\nwant: %v\nhave: %v", post, response)
	}

	// Delete comment error
	postManager.EXPECT().DeleteComment(post.ID.Hex(), comment.ID.Hex()).Return(nil, fmt.Errorf("some error with comment"))

	request = httptest.NewRequest(method, path, nil)
	request = mux.SetURLVars(request, vars)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

type successResponse struct {
	Message string `json:"message"`
}

func TestDelete(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)

	path := fmt.Sprintf("/api/post/%s", post.ID.Hex())
	method := http.MethodDelete
	handler := postHandler.Delete

	// good response
	postManager.EXPECT().Delete(post.ID.Hex()).Return(nil)

	request := httptest.NewRequest(method, path, nil)

	vars := map[string]string{
		"postID": post.ID.Hex(),
	}
	request = mux.SetURLVars(request, vars)

	response := &successResponse{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if response.Message != "success" {
		t.Fatalf(`expected 'success', but was %v`, response.Message)
	}

	// Delete comment error
	postManager.EXPECT().Delete(post.ID.Hex()).Return(fmt.Errorf("some error when try to delete"))

	request = httptest.NewRequest(method, path, nil)

	request = mux.SetURLVars(request, vars)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

func TestGetAllByUser(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)
	post2 := getDefaultPost(author)

	path := "/api/user/" + author.Username
	method := http.MethodGet
	handler := postHandler.GetAllByUser

	// good response
	expected := []*models.Post{post, post2}
	postManager.EXPECT().GetAllByUser(author.Username).Return(expected, nil)

	request := httptest.NewRequest(method, path, nil)

	vars := map[string]string{
		"username": author.Username,
	}
	request = mux.SetURLVars(request, vars)

	response := &[]*models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	for i, post := range *response {
		if !reflect.DeepEqual(post, expected[i]) {
			t.Fatalf(`expected %v, but was %v`, expected[i], post)
		}
	}

	// Get all by user error
	postManager.EXPECT().GetAllByUser(author.Username).Return(nil, fmt.Errorf("some error when try to filter by username"))

	request = httptest.NewRequest(method, path, nil)

	request = mux.SetURLVars(request, vars)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

func TestGetAllByCategory(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)
	post2 := getDefaultPost(author)

	path := "/api/posts/" + post.Category
	method := http.MethodGet
	handler := postHandler.GetAllByCategory

	// good response
	expected := []*models.Post{post, post2}
	postManager.EXPECT().GetAllByCategory(post.Category).Return(expected, nil)

	request := httptest.NewRequest(method, path, nil)

	vars := map[string]string{
		"category": post.Category,
	}
	request = mux.SetURLVars(request, vars)

	response := &[]*models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	for i, post := range *response {
		if !reflect.DeepEqual(post, expected[i]) {
			t.Fatalf(`expected %v, but was %v`, expected[i], post)
		}
	}

	// Get all by category error
	postManager.EXPECT().GetAllByCategory(post.Category).Return(nil, fmt.Errorf("some error when try to filter by category"))

	request = httptest.NewRequest(method, path, nil)

	request = mux.SetURLVars(request, vars)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

func TestGetByID(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)

	path := "/api/posts/" + post.ID.Hex()
	method := http.MethodGet
	handler := postHandler.GetByID

	// good response
	postManager.EXPECT().FindOne(post.ID.Hex()).Return(post, nil)

	request := httptest.NewRequest(method, path, nil)

	vars := map[string]string{
		"postID": post.ID.Hex(),
	}
	request = mux.SetURLVars(request, vars)

	response := &models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if !reflect.DeepEqual(post, response) {
		t.Fatalf(`expected %v, but was %v`, post, response)
	}

	// FindOne error
	postManager.EXPECT().FindOne(post.ID.Hex()).Return(nil, fmt.Errorf("some error when try to find by id"))

	request = httptest.NewRequest(method, path, nil)

	request = mux.SetURLVars(request, vars)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

func TestGetAll(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postManager := NewMockpostManager(ctrl)

	postHandler := &PostHandler{
		Logger:      logger,
		PostManager: postManager,
	}

	author := getDefaultAuthor()
	post := getDefaultPost(author)
	post2 := getDefaultPost(author)

	path := "/api/posts/"
	method := http.MethodGet
	handler := postHandler.GetAll

	// good response
	expected := []*models.Post{post, post2}
	postManager.EXPECT().GetAll().Return(expected, nil)

	request := httptest.NewRequest(method, path, nil)

	response := &[]*models.Post{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err := utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	for i, post := range *response {
		if !reflect.DeepEqual(post, expected[i]) {
			t.Fatalf(`expected %v, but was %v`, expected[i], post)
		}
	}

	// Get all by category error
	postManager.EXPECT().GetAll().Return(nil, fmt.Errorf("some error when try to get all"))

	request = httptest.NewRequest(method, path, nil)

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnprocessableEntity,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}
}

func getDefaultPost(author *models.Author) *models.Post {
	vote := models.Vote{
		Vote:   1,
		UserID: author.ID,
	}
	post := &models.Post{
		Title:            "title",
		Text:             "test text",
		Type:             "text",
		Category:         "music",
		Author:           *author,
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

func getDefaultAuthor() *models.Author {
	author := &models.Author{
		ID:       primitive.NewObjectID(),
		Username: "username",
	}
	return author
}
