package store

import (
	"database/sql"
	"time"
)

func InsertSummaryModel(db *sql.DB, id string) (string, error) {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO summary_models (id, created_at) VALUES (?, ?)`,
		id,
		time.Now().Unix(),
	)
	return id, err
}

func InsertEmbeddingModel(db *sql.DB, id string, embeddingDim int) (string, error) {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO embedding_models (id, embedding_dim, created_at) VALUES (?, ?, ?)`,
		id,
		embeddingDim,
		time.Now().Unix(),
	)
	return id, err
}
