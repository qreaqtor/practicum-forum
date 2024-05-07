package mysql

import (
	"fmt"
	"forum/internal/models"
	"reflect"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	table := "user"
	repoUser := NewUserStorage(db, table)

	testUser := &models.User{
		ID:       "1",
		Password: "password",
		Username: "username",
	}

	// ok query
	mock.
		ExpectExec(fmt.Sprintf("INSERT INTO %s", table)).
		WithArgs(testUser.Username, testUser.ID, testUser.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repoUser.Create(testUser)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// query error
	mock.
		ExpectExec(fmt.Sprintf("INSERT INTO %s", table)).
		WithArgs(testUser.Username, testUser.ID, testUser.Password).
		WillReturnError(fmt.Errorf("bad query"))

	err = repoUser.Create(testUser)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// result error
	mock.
		ExpectExec(fmt.Sprintf("INSERT INTO %s", table)).
		WithArgs(testUser.Username, testUser.ID, testUser.Password).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))

	err = repoUser.Create(testUser)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFindOne(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	table := "user"
	repoUser := NewUserStorage(db, table)

	username := "usertest"

	// good query
	rows := sqlmock.NewRows([]string{"username", "id", "password"})
	expect := []*models.User{
		{
			ID:       "1",
			Password: "password",
			Username: username,
		},
	}
	for _, item := range expect {
		rows = rows.AddRow(item.Username, item.ID, item.Password)
	}

	mock.
		ExpectQuery(fmt.Sprintf("SELECT username, id, password FROM %s WHERE", table)).
		WithArgs(username).
		WillReturnRows(rows)

	item, err := repoUser.FindOne(username)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	// query error
	mock.
		ExpectQuery(fmt.Sprintf("SELECT username, id, password FROM %s WHERE", table)).
		WithArgs(username).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repoUser.FindOne(username)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"username", "id"}).
		AddRow("username", 1)

	mock.
		ExpectQuery(fmt.Sprintf("SELECT username, id, password FROM %s WHERE", table)).
		WithArgs(username).
		WillReturnRows(rows)

	_, err = repoUser.FindOne(username)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}
