package db

import (
	"database/sql"
)

func Connect(dbConnection *string) (*sql.DB, error) {
	db, err := sql.Open("postgres", *dbConnection)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}