package model

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Model struct {
	db *sql.DB
}

// Return new model instance from DB path string.
func NewModel(dbPath string) (*Model, error) {
	db, err := getDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	return &Model{db: db}, nil
}
