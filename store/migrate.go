package store

import "database/sql"

func Migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			path TEXT NOT NULL UNIQUE,
			content_hash TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS summary_models (
			id TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS embedding_models (
			id TEXT PRIMARY KEY,
			embedding_dim INTEGER NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS document_summaries (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			summary_model_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			UNIQUE(document_id, summary_model_id),
			FOREIGN KEY(document_id) REFERENCES documents(id),
			FOREIGN KEY(summary_model_id) REFERENCES summary_models(id)
		)`,
		`CREATE TABLE IF NOT EXISTS document_summary_embeddings (
			summary_id TEXT NOT NULL,
			embedding_model_id TEXT NOT NULL,
			embedding BLOB NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY(summary_id, embedding_model_id),
			FOREIGN KEY(summary_id) REFERENCES document_summaries(id),
			FOREIGN KEY(embedding_model_id) REFERENCES embedding_models(id)
		)`,
		`CREATE TABLE IF NOT EXISTS document_embedding_chunks (
			document_id TEXT NOT NULL,
			embedding_model_id TEXT NOT NULL,
			chunk_index INTEGER NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			content TEXT NOT NULL,
			embedding BLOB NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY(document_id, embedding_model_id, chunk_index),
			FOREIGN KEY(document_id) REFERENCES documents(id),
			FOREIGN KEY(embedding_model_id) REFERENCES embedding_models(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_document_summaries_document ON document_summaries(document_id)`,
		`CREATE INDEX IF NOT EXISTS idx_document_chunks_document ON document_embedding_chunks(document_id)`,
	}

	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}

	return tx.Commit()
}
