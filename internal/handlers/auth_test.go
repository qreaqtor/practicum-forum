package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"forum/internal/handlers/utils"
	"forum/internal/models"
	"testing"

	"github.com/golang/mock/gomock"
)

type authResponse struct {
	Token string `json:"token"`
}

func TestRegister(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authManager := NewMockAuthManager(ctrl)

	userHandler := &UserHandler{
		Logger:      logger,
		AuthManager: authManager,
	}

	path := "/api/register"
	method := http.MethodPost
	handler := userHandler.Register

	// good response
	user := &models.User{
		Password: "password",
		Username: "username",
	}
	expectedToken := "test_token"
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	authManager.EXPECT().Register(user).Return(expectedToken, nil)

	request := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	response := &authResponse{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err = utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if response.Token != expectedToken {
		t.Errorf("\nwant: %v\nhave: %v", expectedToken, response.Token)
	}

	// Register error
	authManager.EXPECT().Register(user).Return("", fmt.Errorf("cant generate token"))

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

	// Validate, marshall, read request
	notValidItem := &models.User{
		Password: "bad",
		Username: "bad",
	}
	err = utils.CheckCasesWithoutManager(handler, path, method, notValidItem)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogin(t *testing.T) {
	logger := slog.New(utils.DummyLogger{})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authManager := NewMockAuthManager(ctrl)

	userHandler := &UserHandler{
		Logger:      logger,
		AuthManager: authManager,
	}

	path := "/api/login"
	method := http.MethodPost
	handler := userHandler.Login

	// good response
	user := &models.User{
		Password: "password",
		Username: "username",
	}
	expectedToken := "test_token"
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	authManager.EXPECT().Login(user).Return(expectedToken, nil)

	request := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	response := &authResponse{}

	test := utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusOK,
		ResponsePtr:    response,
	}

	err = utils.SendTestRequest(test)
	if err != nil {
		t.Fatalf("expected nil, but was %v", err)
	}
	if response.Token != expectedToken {
		t.Errorf("\nwant: %v\nhave: %v", expectedToken, response.Token)
	}

	// Login error
	authManager.EXPECT().Login(user).Return("", fmt.Errorf("bad password"))

	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	test = utils.TestRequest{
		Handler:        handler,
		Request:        request,
		ExpectedStatus: http.StatusUnauthorized,
	}

	err = utils.SendTestRequest(test)
	if err == nil {
		t.Fatal("expected error, but was nil")
	}

	// Validate, marshall, read request
	notValidItem := &models.User{
		Password: "bad",
		Username: "bad",
	}
	err = utils.CheckCasesWithoutManager(handler, path, method, notValidItem)
	if err != nil {
		t.Fatal(err)
	}
}
