package model

import (
	"database/sql"
	"os"
)

const dbSchema = selfSchema + peerSchema

// Get SQL instance from DB path string.
func getDatabase(dbPath string) (*sql.DB, error) {
	_, err := os.Stat(dbPath)
	create := false
	if err != nil {
		if os.IsNotExist(err) {
			// Keep track if db file doesn't exist
			create = true
		} else {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if create {
		// Create tables if new db file
		_, err = db.Exec(dbSchema)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}
