package managers

import (
	"errors"
	"forum/internal/models"
	"reflect"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	secret        = []byte("secret")
	tokenLifespan = 168 // in hours

	errBadPass    = errors.New("invalid password")
	errBadToken   = errors.New("bad token")
	errNoPayload  = errors.New("no payload")
	errBadPayload = errors.New("wrong value type in payload")
)

type sessionRepo interface {
	Set(string, *models.Author) error
	Get(string) (*models.Author, error)
}

type userRepo interface {
	FindOne(string) (*models.User, error)
	Create(*models.User) error
}

type AuthManager struct {
	users    userRepo
	sessions sessionRepo
}

func NewSeesionManager(users userRepo, sessions sessionRepo) *AuthManager {
	return &AuthManager{
		users:    users,
		sessions: sessions,
	}
}

// Выполняет аутентификацию
func (am *AuthManager) Login(loginInput *models.User) (string, error) {
	user, err := am.users.FindOne(loginInput.Username)
	if err != nil {
		return "", err
	}

	userID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return "", nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInput.Password))
	if err != nil {
		return "", errBadPass
	}

	author := &models.Author{ID: userID, Username: user.Username}
	token, err := am.generateToken(author)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Выполняет создание user
func (am *AuthManager) Register(user *models.User) (string, error) {
	userID := primitive.NewObjectID()
	user.ID = userID.Hex()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost) // DefaultCost=10, читал, что сейчас надо побольше, но тут думаю будет норм
	if err != nil {
		return "", err
	}
	user.Password = string(hashedPassword)

	err = am.users.Create(user)
	if err != nil {
		return "", err
	}

	author := &models.Author{ID: userID, Username: user.Username}
	token, err := am.generateToken(author)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Выполняет генерацию и сохранение токена
func (am *AuthManager) generateToken(author *models.Author) (string, error) {
	iat := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": author,
		"iat":  iat.Unix(),
		"exp":  iat.Add(time.Hour * time.Duration(tokenLifespan)).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	err = am.sessions.Set(tokenString, author)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Проверка токена. В горутине осуществляется запрос на получение автора по токену,
// т.к. эта процедура не зависит от извлечения автора из токена
func (am *AuthManager) Check(tokenIn string) (*models.Author, error) {
	errChan := make(chan error)
	authorChan := make(chan *models.Author)
	go func(errChan chan error, authorChan chan *models.Author) {
		author, err := am.sessions.Get(tokenIn)
		if err != nil {
			errChan <- err
		} else {
			authorChan <- author
		}
		close(errChan)
		close(authorChan)
	}(errChan, authorChan)

	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, errBadToken
		}
		return secret, nil
	}
	token, err := jwt.Parse(tokenIn, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errNoPayload
	}

	authorData, ok := payload["user"].(map[string]interface{})
	if !ok {
		return nil, errBadPayload
	}

	idStr, ok := authorData["id"].(string)
	if !ok {
		return nil, errBadPayload
	}

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, errBadPayload
	}

	username, ok := authorData["username"].(string)
	if !ok {
		if !ok {
			return nil, errBadPayload
		}
	}

	authorFromToken := &models.Author{
		ID:       id,
		Username: username,
	}

	select {
	case err := <-errChan:
		return nil, err
	case author := <-authorChan:
		if !reflect.DeepEqual(authorFromToken, author) {
			return nil, errBadToken
		}
		return author, nil
	}
}
