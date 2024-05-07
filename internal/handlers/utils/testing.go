package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

var (
	errExpectedNotNil = errors.New("expected error, but was nil")
)

type TestRequest struct {
	Handler        http.HandlerFunc
	Request        *http.Request
	ResponsePtr    any
	ExpectedStatus int
}

// Проверяет ошибки валидации, маршалинга и чтения
func CheckCasesWithoutManager(handler http.HandlerFunc, path, method string, notValidItem any) error {
	testRequest := TestRequest{
		Handler:     handler,
		ResponsePtr: nil,
	}

	// Read request body error
	request := httptest.NewRequest(method, path, nil)

	testRequest.Request = request
	testRequest.ExpectedStatus = http.StatusBadRequest

	err := SendTestRequest(testRequest)
	if err == nil {
		return errExpectedNotNil
	}

	// Unmarshall error
	request = httptest.NewRequest(method, path, bytes.NewBuffer([]byte("not JSON")))
	request.Header.Set("Content-Type", "application/json")

	testRequest.Request = request
	testRequest.ExpectedStatus = http.StatusUnprocessableEntity

	err = SendTestRequest(testRequest)
	if err == nil {
		return errExpectedNotNil
	}

	// Valid error
	body, err := json.Marshal(notValidItem)
	if err != nil {
		return err
	}
	request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	testRequest.Request = request

	err = SendTestRequest(testRequest)
	if err == nil {
		return errExpectedNotNil
	}

	return nil
}

func SendTestRequest(test TestRequest) error {
	w := httptest.NewRecorder()

	test.Handler(w, test.Request)

	response := w.Result()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != test.ExpectedStatus {
		return fmt.Errorf("error: expected response status %d, got %d: %s", test.ExpectedStatus, response.StatusCode, responseBody)
	}
	return json.Unmarshal(responseBody, test.ResponsePtr)
}
