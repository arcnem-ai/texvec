package store

import (
	"database/sql"

	"github.com/arcnem-ai/texvec/config"
	_ "github.com/tursodatabase/go-libsql"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("libsql", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	return db, nil
}

func GetDB() (*sql.DB, error) {
	dbPath, err := config.DBPath()
	if err != nil {
		return nil, err
	}

	return Open(dbPath)
}
