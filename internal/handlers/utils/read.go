package utils

import (
	"errors"
	"io"
	"net/http"
)

var (
	errUnknownPayload = errors.New("unknown payload")
)

// Выполняет проверку заголовка Content-type и читает body.
func ReadRequestBody(r *http.Request) ([]byte, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, errUnknownPayload
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()

	return body, nil
}
