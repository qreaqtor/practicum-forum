package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"forum/internal/handlers/utils"
	"forum/internal/models"
)

type AuthManager interface {
	Login(*models.User) (string, error)
	Register(*models.User) (string, error)
}

type UserHandler struct {
	Logger      *slog.Logger
	AuthManager AuthManager
}

// Хендлер аутентификации
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(uh.Logger, r.URL.Path, r.Method)

	body, err := utils.ReadRequestBody(r)
	if err != nil {
		msg.Set(err.Error(), http.StatusBadRequest)
		utils.WriteError(w, msg)
		return
	}

	loginInput := &models.User{}
	err = json.Unmarshal(body, loginInput)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	err = utils.ValidateStruct(loginInput)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	token, err := uh.AuthManager.Login(loginInput)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnauthorized)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, map[string]interface{}{
		"token": token,
	})
}

// Хендлер регистрации
func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	msg := utils.NewLogMsg(uh.Logger, r.URL.Path, r.Method)

	body, err := utils.ReadRequestBody(r)
	if err != nil {
		msg.Set(err.Error(), http.StatusBadRequest)
		utils.WriteError(w, msg)
		return
	}

	user := &models.User{}
	err = json.Unmarshal(body, user)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	err = utils.ValidateStruct(user)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	token, err := uh.AuthManager.Register(user)
	if err != nil {
		msg.Set(err.Error(), http.StatusUnprocessableEntity)
		utils.WriteError(w, msg)
		return
	}

	msg.Set("success", http.StatusOK)
	utils.WriteData(w, msg, map[string]interface{}{
		"token": token,
	})
}
