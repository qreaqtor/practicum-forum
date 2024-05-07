package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func GetConnection(dbUser, dbPassword, addr, dbName string) (*sql.DB, error) {
	// основные настройки к базе
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?", dbUser, dbPassword, addr, dbName)
	// указываем кодировку
	dsn += "&charset=utf8"
	// отказываемся от prapared statements
	// параметры подставляются сразу
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
