package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/internal/models"
)

var (
	errUserExists = errors.New("this username already exists")
)

type userStorage struct {
	db    *sql.DB
	table string
}

func NewUserStorage(db *sql.DB, userTable string) *userStorage {
	return &userStorage{
		db:    db,
		table: userTable,
	}
}

// Возрващает первый элемент по username
func (us *userStorage) FindOne(username string) (*models.User, error) {
	user := &models.User{}
	query := fmt.Sprintf("SELECT username, id, password FROM %s WHERE username = ?", us.table)
	row := us.db.QueryRow(
		query,
		username,
	)
	err := row.Scan(&user.Username, &user.ID, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Создает пользователя
func (us *userStorage) Create(user *models.User) error {
	query := fmt.Sprintf("INSERT INTO %s (username, id, password) VALUES (?, ?, ?)", us.table)
	result, err := us.db.Exec(
		query,
		user.Username,
		user.ID,
		user.Password,
	)
	if err != nil {
		return errUserExists
	}
	affected, err := result.RowsAffected()
	if affected == 0 {
		return errUserExists
	}
	return err
}
